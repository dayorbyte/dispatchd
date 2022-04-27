from collections import defaultdict
import sys
import subprocess
import os

PREFIX = 'github.com/dayorbyte/dispatchd/'

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
  output_dir = os.path.join(gopath, 'cover')
  if not os.path.exists(output_dir):
    os.makedirs(output_dir)
  package_dir = os.path.join(gopath, 'src', PREFIX)
  prefix = 'github.com/dayorbyte/dispatchd/'
  test_packages = subprocess.check_output([
    'go',
    'list',
    '{}...'.format(prefix),
  ]).split('\n')
  test_packages = [t.strip().replace(prefix, '') for t in test_packages]
  test_packages = [t for t in test_packages if t]

  cover_names = []
  for pkg in test_packages:
    cover_name = os.path.join(output_dir, '{}.cover'.format(pkg))
    cover_names.append((pkg, cover_name))
    test_target = os.path.join(PREFIX, pkg)
    cmd = [
      'go',
      'test',
      '-coverpkg=github.com/dayorbyte/dispatchd/...',
      '-coverprofile={}'.format(cover_name),
      test_target,
    ]
    subprocess.check_call(cmd)

  merge_call = [os.path.join(gopath, 'bin', 'gocovmerge')]
  merge_call.extend([n for _, n in cover_names if os.path.exists(n)])
  output = subprocess.check_output(merge_call)
  all = os.path.join(output_dir, 'all.cover')
  with open(all, 'w') as f:
    f.write(output)
  cover_summary([all])

def nocover(file, line):
  with open(os.path.join(os.environ['GOPATH'], 'src', file)) as inf:
    lines = inf.readlines()
    if 'pragma: nocover' in lines[int(line)-1]:
      return True
    return False

def count_lines(file):
  with open(os.path.join(os.environ['GOPATH'], 'src', file)) as inf:
    return len(inf.readlines())

def cover_summary(cover_names):
  print Colors.BLUE, '=====  Missing Coverage =====', Colors.ENDC
  count = defaultdict(list)
  missing = defaultdict(list)
  total_missing = 0
  total_lines = 0
  seen = set()
  for name in cover_names:
    with open(name) as inf:
      for line in inf.readlines():
        line = line.strip()
        if line[-1] != '0':
          continue
        full_file, _, report = line.partition(':')
        file = full_file.replace(PREFIX, '')
        if 'pb.go' in file or 'generated' in file or 'testlib' in file:
          continue
        range, _, _ = report.partition(' ')
        first, _, second = range.partition(',')
        first_line, _, _ = first.partition('.')
        second_line, _, _ = second.partition('.')
        range = (int(first_line), int(second_line))
        if full_file not in seen:
          total_lines += count_lines(full_file)
          seen.add(full_file)
        total_missing += range[1] - range[0] + 1
        if nocover(full_file, first_line):
          continue
        missing[file].append(range)


  for file, ranges in sorted(missing.items()):
    if ranges is not None:
      ranges = sorted(ranges)
      ranges = ['{}-{}'.format(x,y) for x, y in ranges]
      if len(ranges) == 0:
        print Colors.GREEN, file+':', Colors.ENDC, 'Full coverage!'
      print Colors.RED, file+':', Colors.ENDC, ', '.join(ranges)
    else:
      print Colors.RED, file+':', Colors.ENDC, 'No coverage'
  print Colors.RED, 'Remaining lines on files without full coverage:', total_missing, '/', total_lines, Colors.ENDC


if __name__ == '__main__':
  main(sys.argv[1:])
