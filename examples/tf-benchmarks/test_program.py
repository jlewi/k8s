"""Simple program to test launcher"""
import logging
import time
if __name__ == "__main__":

  logging.getLogger().setLevel(logging.INFO)
  print("Hello from print in the test program.")
  logging.info("Log from test_proram")
  while True:
    print("Test program print sleep")
    logging.info("Test program log sleep")
    time.sleep(300)