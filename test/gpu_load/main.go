// e2e provides an E2E test for TfJobs.
//
// The test creates TfJobs and runs various checks to ensure various operations work as intended.
// The test is intended to run as a helm test that ensures the TfJob operator is working correctly.
// Thus, the program returns non-zero exit status on error.
//
// TODO(jlewi): Do we need to make the test output conform to the TAP(https://testanything.org/)
// protocol so we can fit into the K8s dashboard
//
package main

import (
	"flag"
	"fmt"
	"github.com/gogo/protobuf/proto"
	log "github.com/golang/glog"
	k8s_errors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/pkg/api/v1"
	"mlkube.io/pkg/spec"
	"mlkube.io/pkg/util"
	"mlkube.io/pkg/util/k8sutil"
	"strings"
	"time"
)

const (
	Namespace = "default"
)

var (
	image   = flag.String("image", "gcr.io/cloud-ml-dev/tf_smoke_cmle:latest", "The Docker image to use with the TfJob.")
	repeats = flag.Int("repeats", 200, "Number of times to Repeat the test.")
)

func run() error {
	kubeCli := k8sutil.MustNewKubeClient()
	tfJobClient, err := k8sutil.NewTfJobClient()
	if err != nil {
		return err
	}

	name := "gpu-load-test-job-" + util.RandString(4)

	original := &spec.TfJob{
		Metadata: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				"test.mlkube.io": "",
			},
		},
		Spec: spec.TfJobSpec{
			ReplicaSpecs: []*spec.TfReplicaSpec{
				{
					Replicas:      proto.Int32(1),
					TfPort:        proto.Int32(2222),
					TfReplicaType: spec.MASTER,
					Template: &v1.PodTemplateSpec{
						Spec: v1.PodSpec{
							Containers: []v1.Container{
								{
									Name:  "tensorflow",
									Image: *image,

                  Args: []string {
                    "--gpu",
                  },
									Env: []v1.EnvVar{
										{
											Name:  "LD_LIBRARY_PATH",
											Value: "/usr/local/cuda/lib64",
										},
									},
									SecurityContext: &v1.SecurityContext{
										Privileged: proto.Bool(true),
									},
									VolumeMounts: []v1.VolumeMount{
										{
											Name:      "dev-nvidia",
											MountPath: "/dev/nvidia0",
										},
										{
											Name:      "dev-nvidiactl",
											MountPath: "/dev/nvidiactl",
										},
										{
											Name:      "dev-nvidia-uvm",
											MountPath: "/dev/nvidia-uvm",
										},
									},
								},
							},
              Volumes: []v1.Volume{
                {
                  Name: "dev-nvidia",
                  VolumeSource: v1.VolumeSource{
                    HostPath: &v1.HostPathVolumeSource{
                      Path: "/dev/nvidia0",
                    },
                  },
                },
                {
                  Name: "dev-nvidiactl",
                  VolumeSource: v1.VolumeSource{
                    HostPath: &v1.HostPathVolumeSource{
                      Path: "/dev/nvidiactl",
                    },
                  },
                },
                {
                  Name: "dev-nvidia-uvm",
                  VolumeSource: v1.VolumeSource{
                    HostPath: &v1.HostPathVolumeSource{
                      Path: "/dev/nvidia-uvm",
                    },
                  },
                },
              },
							RestartPolicy: v1.RestartPolicyOnFailure,
						},
					},
				},
			},
		},
	}

	_, err = tfJobClient.Create(Namespace, original)

	if err != nil {
		return err
	}

	// Wait for the job to complete for up to 2 minutes.
	var tfJob *spec.TfJob
	for endTime := time.Now().Add(2 * time.Minute); time.Now().Before(endTime); {
		tfJob, err = tfJobClient.Get(Namespace, name)
		if err != nil {
			log.Warningf("There was a problem getting TfJob: %v; error %v", name, err)
		}

		if tfJob.Status.State == spec.StateSucceeded || tfJob.Status.State == spec.StateFailed {
			break
		}
		log.Infof("Waiting for job %v to finish", name)
		time.Sleep(5 * time.Second)
	}

	if tfJob == nil {
		return fmt.Errorf("Failed to get TfJob %v", name)
	}

	if tfJob.Status.State != spec.StateSucceeded {
		// TODO(jlewi): Should we clean up the job.
		return fmt.Errorf("TfJob %v did not succeed;\n %v", name, util.Pformat(tfJob))
	}

	if tfJob.Spec.RuntimeId == "" {
		return fmt.Errorf("TfJob %v doesn't have a RuntimeId", name)
	}

	// Loop over each replica and make sure the expected resources were created.
	for _, r := range original.Spec.ReplicaSpecs {
		baseName := strings.ToLower(string(r.TfReplicaType))

		for i := 0; i < int(*r.Replicas); i += 1 {
			jobName := fmt.Sprintf("%v-%v-%v", baseName, tfJob.Spec.RuntimeId, i)

			_, err := kubeCli.BatchV1().Jobs(Namespace).Get(jobName, metav1.GetOptions{})

			if err != nil {
				return fmt.Errorf("Tfob %v did not create Job %v for ReplicaType %v Index %v", name, jobName, r.TfReplicaType, i)
			}
		}
	}

  // DO NOT Submit skip delete
  return nil

	// Delete the job and make sure all subresources are properly garbage collected.
	if err := tfJobClient.Delete(Namespace, name); err != nil {
		log.Fatal("Failed to delete TfJob %v; error %v", name, err)
	}

	// Define sets to keep track of Job controllers corresponding to Replicas
	// that still exist.
	jobs := make(map[string]bool)

	// Loop over each replica and make sure the expected resources are being deleted.
	for _, r := range original.Spec.ReplicaSpecs {
		baseName := strings.ToLower(string(r.TfReplicaType))

		for i := 0; i < int(*r.Replicas); i += 1 {
			jobName := fmt.Sprintf("%v-%v-%v", baseName, tfJob.Spec.RuntimeId, i)

			jobs[jobName] = true
		}
	}

	// Wait for all jobs to be deleted.
	for endTime := time.Now().Add(5 * time.Minute); time.Now().Before(endTime) && len(jobs) > 0; {
		for k := range jobs {
			_, err := kubeCli.BatchV1().Jobs(Namespace).Get(k, metav1.GetOptions{})
			if k8s_errors.IsNotFound(err) {
				// Deleting map entry during loop is safe.
				// See: https://stackoverflow.com/questions/23229975/is-it-safe-to-remove-selected-keys-from-golang-map-within-a-range-loop
				delete(jobs, k)
			} else {
				log.Infof("Job %v still exists", k)
			}
		}

		if len(jobs) > 0 {
			time.Sleep(5 * time.Second)
		}
	}

	if len(jobs) > 0 {
		return fmt.Errorf("Not all Job controllers were successfully deleted.")
	}
	return nil
}

func main() {
	flag.Parse()

	for i := 0; i < *repeats; i++ {
		err := run()

		// Generate TAP (https://testanything.org/) output
		fmt.Println("1..1")
		if err == nil {
			fmt.Println("ok 1 - Successfully ran TfJob")
		} else {
			fmt.Println("not ok 1 - Running TfJob failed %v", err)
		}
	}
}
