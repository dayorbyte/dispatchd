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
    f.write('package amqp\n\n')
    f.write('''import (
  //"encoding/binary"
  "io"
  "errors"
  //"bytes"
  "strconv"
)\n\n\n''')
    methods = []
    for child in root:
      if child.tag != 'class':
        continue
      methods.extend(handle_class(f, child, domains))
    handle_method_frame_reader(f, methods)

def handle_class(f, node, domains):
  cls_name = normalize_name(node.attrib['name'])
  cls_index = node.attrib['index']
  comment_header(f, cls_name, big=True)
  f.write('var ClassId{} uint16 = {}\n'.format(cls_name, cls_index))
  methods = []
  for child in node:
    if child.tag != 'method':
      continue
    methods.append(handle_method(f, child, cls_name, cls_index, domains))
  return methods

def handle_method(f, node, cls_name, cls_index, domains):
  method_name = normalize_name(node.attrib['name'])
  method_index = node.attrib['index']
  comment_header(f, '{} - {}'.format(cls_name, method_name))
  f.write('var MethodId{}{} uint16 = {}\n'.format(cls_name, method_name, method_index))
  fields = []
  for child in node:
    if child.tag == 'field':
      try:
        domain = child.attrib.get('domain')
        if not domain:
          domain = child.attrib.get('type')
        type = domains[domain]

        fields.append(dict(
          name=normalize_name(child.attrib['name']),
          type=type,
          serializer=normalize_name(domain)
        ))
      except:
        print child, child.attrib['name'], child.attrib
        raise
  struct_name = '{}{}'.format(cls_name, method_name)
  handle_method_struct(f, struct_name, fields, cls_index, method_index)
  handle_method_reader(f, struct_name, fields)
  handle_method_writer(f, struct_name, fields, cls_index, method_index)
  return (cls_index, cls_name, method_index, struct_name)


def handle_method_struct(f, struct_name, fields, cls_index, method_index):
  f.write('''type {} struct {{\n'''.format(struct_name))
  for field in fields:
    f.write('  {} {}\n'.format(field['name'], field['type']))
  f.write('}\n\n')
  f.write('''
func (f* {}) MethodIdentifier() (uint16, uint16) {{
  return {}, {}
}}\n\n'''.format(struct_name, cls_index, method_index))

def handle_method_reader(f, struct_name, fields):
  f.write('''func (f *{}) Read(reader io.Reader) (err error) {{\n'''.format(struct_name))
  bits = 0
  for field in fields:
    if field['type'] == 'bool':
      if bits == 0:
        read_bits(f)
      f.write('''  f.{name} = (bits & (1 << {bits}) > 0)\n\n'''.format(bits=bits, **field))
      bits += 1
    else:
      bits = 0
      f.write('''  f.{name}, err = Read{serializer}(reader)
  if err != nil {{
    return errors.New("Error reading field {name}")
  }}\n\n'''.format(**field))
  f.write('  return\n')
  f.write('}\n')

def read_bits(f):
  f.write('''
  bits, err := ReadOctet(reader)
  if err != nil {{
    return errors.New("Error reading field {name}")
  }}\n\n
''')

def handle_method_writer(f, struct_name, fields, cls_index, method_index):
  f.write('''func (f *{}) Write(writer io.Writer) (err error) {{\n'''.format(struct_name))
  for name in [cls_index, method_index]:
    f.write('''  if err = WriteShort(writer, {}); err != nil {{
    return err
  }}\n'''.format(name))
  bits = 0
  for field in fields:
    if field['type'] == 'bool':
      if bits == 0:
        f.write('  var bits byte\n')
      f.write('''  if f.{name} {{\n    bits |= 1 << {bits}\n  }}\n'''.format(bits=bits, **field))
      bits += 1
    else:
      if bits > 0:
        write_bits(f)
        bits = 0
      f.write('''  err = Write{serializer}(writer, f.{name})
  if err != nil {{
    return errors.New("Error writing field {name}")
  }}\n\n'''.format(**field))
  if bits > 0:
    write_bits(f)
  f.write('  return\n')
  f.write('}\n')

def write_bits(f):
  f.write('''  err = WriteOctet(writer, bits)
  if err != nil {{
    return errors.New("Error writing bit fields")
  }}\n\n''')

def comment_header(f, name, big=False):
  f.write('\n// ' + '*' * 70 + '\n')
  if big:
    f.write('//\n//\n')
  f.write('//' + ' ' * 20 + name + '\n')
  if big:
    f.write('//\n//\n')
  f.write('// ' + '*' * 70 + '\n\n')

def handle_domains(node):
  domains = {}
  with open(DOMAIN_FILE, 'w') as f:
    f.write('package amqp\n\n')
    for child in node:
      if child.tag == 'domain':
        name = child.attrib['name']
        type = child.attrib['type']
        domains[name] = amqp_to_go_types[type]
        if type != name:
          f.write('var Read{} = Read{}\n'.format(normalize_name(name), normalize_name(type)))
          f.write('var Write{} = Write{}\n'.format(normalize_name(name), normalize_name(type)))

  return domains

def handle_method_frame_reader(f, methods):
  methods = sorted(methods)
  # open fn body, open switch
  f.write('''func ReadMethod(reader io.Reader) (MethodFrame, error) {
  classIndex, err := ReadShort(reader)
  if err != nil {
    return nil, err
  }
  methodIndex, err := ReadShort(reader)
  if err != nil {
    return nil, err
  }
  switch {
''')
  last_class = -1
  for cls_index, cls_name, method_index, struct_name in methods:
    if cls_index != last_class:
      # close inner switch
      if last_class != -1:
        f.write('    }\n')
      f.write('''    // {cls_name}
    case classIndex == {cls_index}:
      switch {{\n'''.format(**vars()))
    f.write('''      case methodIndex == {method_index}: // {struct_name}
        var method = &{struct_name}{{}}
        err = method.Read(reader)
        if err != nil {{
          return nil, err
        }}
        return method, nil\n'''.format(**vars()))
    last_class = cls_index
  # close last inner switch, close switch
  f.write('    }\n  }\n')
  # close fn body
  f.write('''  return nil, errors.New(
    "Bad method or class Id! classId:" +
    strconv.FormatUint(uint64(classIndex), 10) +
    " methodIndex: " +
    strconv.FormatUint(uint64(methodIndex), 10))\n''')
  f.write('}')

def normalize_name(name):
  return ''.join([w.capitalize() for w in name.split('-')])

if __name__ == '__main__':
  tree = ET.parse('amqp0-9-1.extended.xml')
  root = tree.getroot()

  domains = handle_domains(root)
  handle_constants(root)
  methods = handle_classes(root, domains)

