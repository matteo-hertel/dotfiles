#!/usr/bin/env python2
import os
import errno
ROOT=os.path.dirname(os.path.realpath(__file__))
HOME=os.path.expanduser('~')
DIST= HOME

def mkdir_p(path):
    try:
        os.makedirs(path)
    except OSError as exc:  # Python >2.5
        if exc.errno == errno.EEXIST and os.path.isdir(path):
            pass
        else:
            raise

mkdir_p(DIST);

for root, dirs, files in os.walk(ROOT):
    for file in files:
        if file.endswith(".link"):
             symlink_root = root.replace(ROOT, DIST);
             mkdir_p(symlink_root);
             os.symlink(os.path.join(root, file), os.path.join(symlink_root, file.replace(".link", "")));
             print "Created", os.path.join(symlink_root, file.replace(".link", ""))
