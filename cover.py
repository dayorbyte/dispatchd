from collections import defaultdict
import sys
import subprocess
import os

PREFIX = 'github.com/jeffjenkins/mq/'

class Colors(object):
  PURPLE = '\033[95m'
  BLUE = '\033[94m'
  GREEN = '\033[92m'
  YELLOW = '\033[93m'
  RED = '\033[91m'
  ENDC = '\033[0m'
  BOLD = '\033[1m'
  UNDERLINE = '\033[4m'

def main(args):
  # Gather packages
  gopath = os.environ['GOPATH']
  package_dir = os.path.join(gopath, 'src', PREFIX)

  test_packages = []
  for path in os.listdir(package_dir):
    if path in ('.git','scripts', 'static'):
      continue
    if os.path.isdir(path):
      test_packages.append(path)


  cover_names = []
  for pkg in test_packages:
    cover_name = 'cover-{}.cover'.format(pkg)
    cover_names.append((pkg, cover_name))
    test_target = os.path.join(PREFIX, pkg)
    subprocess.check_call([
      'go',
      'test',
      '-coverprofile={}'.format(cover_name),
      test_target,
    ])
  cover_summary(cover_names)

def nocover(file, line):
  with open(os.path.join(os.environ['GOPATH'], 'src', file)) as inf:
    lines = inf.readlines()
    if 'pragma: nocover' in lines[int(line)]:
      return True
    return False

def cover_summary(cover_names):
  print Colors.BLUE, '=====  Missing Coverage =====', Colors.ENDC
  count = defaultdict(list)
  missing = defaultdict(list)
  for pkg, name in cover_names:
    if not os.path.exists(name):
      missing[pkg + '/*.go'] = ['No coverage']
      continue
    with open(name) as inf:
      for line in inf.readlines():
        line = line.strip()
        if line[-1] != '0':
          continue
        full_file, _, report = line.partition(':')
        file = full_file.replace(PREFIX, '')
        if 'pb.go' in file:
          continue
        range, _, _ = report.partition(' ')
        first, _, second = range.partition(',')
        first_line, _, _ = first.partition('.')
        second_line, _, _ = second.partition('.')
        range = '{}-{}'.format(first_line, second_line)
        if nocover(full_file, first_line):
          continue
        missing[file].append(range)

  for file, ranges in missing.iteritems():
    if len(ranges) > 20:
      ranges = ranges[:20]
      ranges.append('Report Truncated.')
    if len(ranges) == 0:
      print Colors.GREEN, file+':', Colors.ENDC, 'Full coverage!'
    print Colors.RED, file+':', Colors.ENDC, ', '.join(ranges)




if __name__ == '__main__':
  main(sys.argv[1:])