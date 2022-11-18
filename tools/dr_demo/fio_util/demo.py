import os
import sys
import time
import argparse
from helperlib import *

config = read_config()
MAX_PHASE_ID= 50

def fio_demo():
    print("DR demo script\n\n")

    phase_id = eval(config['PHASE_ID']['ID'])
    runtime = config['FIO']['RUNTIME']
    bs = config['FIO']['BS']
    queue = config['FIO']['IO_DEPTH']
    size = config['FIO']['SIZE']
    iops = config['FIO']['IOPS']
    dir_path = config['FIO']['DIRECTORY']
    jobs = config['FIO']['NUMJOBS']
    rw = config['FIO']['OP']
    rwmix = config['FIO']['RW_MIX_READ']

    os.system('clear')

    while(int(phase_id) < MAX_PHASE_ID):
         print("PHASE ID %s\n\n" %phase_id)
         print("Running fio with phase id = %s "%phase_id)
         fio_output = run_fio(runtime,  bs, queue, size, iops, jobs, dir_path,phase_id, rw, rwmix)
         print("fio_output : %s \n\n"%fio_output)
         phase_id = create_file_with_phaseid(phase_id, fio_output)

         update_phase_id(phase_id)
         time.sleep(2)
         sys.stderr.write("\x1b[2J\x1b[H")  # Escape sequence
         os.system('clear')

if __name__ == '__main__':
    fio_demo()

