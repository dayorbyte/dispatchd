import os
from mako.template import Template
from mako.lookup import TemplateLookup
from mako import exceptions
import xml.etree.ElementTree as ET

CONSTANTS_FILE = 'amqp/constants_generated.go'
PROTOCOL_FILE = 'amqp/protocol_generated.go'
DOMAIN_FILE = 'amqp/domains_generated.go'

amqp_to_go_types = {
  'bit'   : 'bool',
  'octet' : 'byte',
  'short' : 'uint16',
  'long'  : 'uint32',
  'longlong': 'uint64',
  'table' : 'Table',
  'timestamp' : 'uint64',
  'shortstr' : 'string',
  'longstr' : '[]byte',
}

def handle_constants(root):
  with open(CONSTANTS_FILE, 'w') as f:
    f.write('package amqp\n\n')
    # Manual constants
    f.write('''var MaxShortStringLength uint8 = 255\n''')
    # Protocol constants
    for child in root:
      if child.tag != 'constant':
        continue
      handle_constant(f, child)

def handle_constant(f, constant):
  for child in constant:
    if child.tag == 'doc':
      f.write('\n')
      for line in child.text.split('\n'):
        if line.strip():
          f.write('// {}\n'.format(line.strip()))
  name = constant.attrib['name']
  value = constant.attrib['value']
  name = normalize_name(name)
  f.write('var {} = {}\n'.format(name, value))

def handle_classes(root, domains):
  with open(PROTOCOL_FILE, 'w') as f:
    f.write(render('protocol', root=root, domains=domains))

def handle_domains(node):
  domains = {}
  for child in node:
    if child.tag == 'domain':
      name = child.attrib['name']
      type = child.attrib['type']
      domains[name] = Domain(name, type)

  with open(DOMAIN_FILE, 'w') as f:
    f.write(render('domains', domains=domains.values()))

  return domains

class Domain(object):
  def __init__(self, name, type):
    self.name = name
    self.name_normalized = normalize_name(name)
    self.amqp_type = type
    self.go_type = amqp_to_go_types[type]
    self.custom = name != self.amqp_type

def normalize_name(name):
  return ''.join([w.capitalize() for w in name.split('-')])

def render(name, **kwargs):
  path = os.path.join(os.path.abspath(os.path.dirname(__file__)))
  lookup = TemplateLookup(directories=[path])
  try:
    template = lookup.get_template(name + '.mako')
    return template.render(**kwargs)
  except:
    print(exceptions.text_error_template().render())
    raise

if __name__ == '__main__':
  tree = ET.parse('amqp0-9-1.extended.xml')
  root = tree.getroot()

  domains = handle_domains(root)
  handle_constants(root)
  methods = handle_classes(root, domains)

