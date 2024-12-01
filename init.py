#!/usr/bin/env python2
import os
import errno
ROOT=os.path.dirname(os.path.realpath(__file__))
HOME=os.path.expanduser('~')
DIST= HOME

def mkdir_p(path):
    try:
        os.makedirs(path)
    except OSError as exc:
        if exc.errno == errno.EEXIST and os.path.isdir(path):
            pass
        else:
            raise

def symlink_force(target, link_name):
    try:
        print("Creating", target)
        os.symlink(target, link_name)
    except OSError as e:
        if e.errno == errno.EEXIST:
            os.remove(link_name)
            os.symlink(target, link_name)
            print ("Created", target)

        else:
            raise e
def process_link(name, root):
    if name.endswith(".link"):
           symlink_root = root.replace(ROOT, DIST)
           mkdir_p(symlink_root)
           symlink_force(os.path.join(root, name), os.path.join(symlink_root, name.replace(".link", "")))
 
mkdir_p(DIST)

for root, dirs, files in os.walk(ROOT):
   if(root == ROOT):
      for dir in dirs:
         process_link(dir, root)
   for file in files:
        process_link(file, root)

os.symlink(os.path.join(ROOT, '.config/nvim'), os.path.join(HOME, '.config/nvim'))
