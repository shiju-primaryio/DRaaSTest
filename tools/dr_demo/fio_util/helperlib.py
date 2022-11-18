import os
import configparser
import time
import subprocess
import sys

config_file = "demo.conf"

def read_config():
    """
	Read config params from config file
	"""
    config = configparser.ConfigParser()
    config.sections()
    config.optionxform = str
    config.read(config_file)
    return config

def create_file_with_phaseid(phase_id, fio_output):
    """
    create file with phaseid
    """
    text = fio_output
    file_path = os.getcwd()
    txtfile = file_path + '/' + str(phase_id) + '.txt'

    if not os.path.exists(txtfile):
        phase_id += 1
        text_file = open(txtfile, "a")
        text_file.write(text)
        text_file.close()
    elif os.path.exists(txtfile) and phase_id >= 1:
        phase_id += 1
        text_file1 = open(file_path + '/' + str(phase_id) + '.txt', "a")
        text_file1.write(text)
        text_file1.close()

    return phase_id

def run_fio(runtime, bs, iodepth, size, iops, jobs, dir_path, filename, rw, rwmix):
    """
    Run fio with configurable params
    """
    cmd = "fio --name=random --ioengine=libaio --time_based --iodepth={} --norandommap --refill_buffers --buffer_compress_percentage=30 --group_reporting --gtod_reduce=1 --stonewall --rw={} --bs={} --direct=1 \
 --rate_iops={} --size={} --numjobs={} --randrepeat=0 --directory={} --filename={} --runtime={}".format(iodepth, rw, bs, iops, size, jobs, dir_path, filename, runtime)

    print(cmd)
    stream = os.popen(cmd)
    output = stream.read()
    return output

def update_phase_id(phase_id):
    """
    update phase id in config
    """
    config = read_config()
    print("Updating PHASE ID in conf")
    config.set('PHASE_ID','ID',str(phase_id))
    with open("dr_config.conf",'w') as c:
        config.write(c)
    time.sleep(5)

