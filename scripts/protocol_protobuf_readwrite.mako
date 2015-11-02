<%

# Global State
ALL_METHODS = []

def normalize_field(name):
  return ''.join([w.capitalize() for w in name.split('-')])

def normalize_struct(name):
  return ''.join([w.capitalize() for w in name.split('-')])

def pointer_star(field):
  ## print field['type'].protobuf_type
  if field['type'].protobuf_type in ('Table', 'bytes', 'string'):
    return ''
  ## print field
  return '*'

def iter_tag(node, name):
  for child in node:
    if child.tag == name:
      yield child
%>package amqp

import (
  "io"
  "errors"
  "fmt"
)

% for cls_node in iter_tag(root, 'class'):
<%
  cls_name = normalize_struct(cls_node.attrib['name'])
  cls_index = cls_node.attrib['index']
%>
${comment_header(cls_name, big=True)}
${class_body(cls_node, cls_name, cls_index)}
  % for method_node in iter_tag(cls_node, 'method'):
${method(method_node, cls_name, cls_index)}
  % endfor

% endfor

${read_method()}

<%def name="class_body(cls_node, cls_name, cls_index)"><%
class_fields = list(iter_tag(cls_node, 'field'))
if not class_fields:
  return ''
%>

% for index, field in enumerate(class_fields):
<% hex = '%04x' % (0 | 1 << (15 - index))
%>
var Mask${normalize_field(field.attrib['name'])} uint16 = 0x${hex}
% endfor

</%def>

<%def name="read_method()">
<%
last_class = -1
%>
func ReadMethod(reader io.Reader) (MethodFrame, error) {
  classIndex, err := ReadShort(reader)
  if err != nil {
    return nil, err
  }
  methodIndex, err := ReadShort(reader)
  if err != nil {
    return nil, err
  }
  switch {
% for cls_index, cls_name, method_index, struct_name in sorted(ALL_METHODS):
    % if cls_index != last_class:
      % if last_class != -1:
    }
      % endif
    // ${cls_name}
    case classIndex == ${cls_index}:
      switch {
    % endif
      case methodIndex == ${method_index}: // ${struct_name}
        var method = &${struct_name}{}
        err = method.Read(reader)
        if err != nil {
          return nil, err
        }
        return method, nil

    <% last_class = cls_index %>
% endfor
    }
  }
  return nil, errors.New(fmt.Sprintf("Bad method or class Id! classId: %d, methodIndex: %d", classIndex, methodIndex))

}

</%def>

<%def name="comment_header(name, big=False)">
% if big:
// **********************************************************************
//
//
//                    ${name}
//
//
// **********************************************************************
% else:
// **********************************************************************
//                    ${name}
// **********************************************************************
% endif
</%def>

<%def name="method(method_node, cls_name, cls_index)">
<%
method_name = normalize_field(method_node.attrib['name'])
method_index = method_node.attrib['index']
struct_name = '{}{}'.format(cls_name, method_name)
ALL_METHODS.append((cls_index, cls_name, method_index, struct_name))
fields = []

for field in iter_tag(method_node, 'field'):
  try:
    domain = field.attrib.get('domain')
    if not domain:
      domain = field.attrib.get('type')
    type = domains[domain]

    fields.append(dict(
      name=normalize_field(field.attrib['name']),
      type=type,
      serializer=normalize_field(domain)
    ))
  except:
    print field, field.attrib['name'], field.attrib
    raise

%>
${comment_header('{} - {}'.format(cls_name, method_name))}
var MethodId${cls_name}${method_name} uint16 = ${method_index}

${method_reader(struct_name, fields)}
${method_writer(struct_name, fields, cls_index, method_index)}


</%def>

<%def name="method_reader(struct_name, fields)"><%
bits = 0
%>
func (f *${struct_name}) Read(reader io.Reader) (err error) {
% for field in fields:
  % if field['type'].go_type == 'bool':
    % if bits == 0:
  bits, err := ReadOctet(reader)
  if err != nil {
    return errors.New("Error reading field ${field['name']}" + err.Error())
  }
    % endif

  f.${field['name']} = (bits & (1 << ${bits}) > 0)
<% bits += 1 %>
  % else:
<% bits = 0 %>
  ${pointer_star(field)}f.${field['name']}, err = Read${field['serializer']}(reader)
  % endif
  if err != nil {
    return errors.New("Error reading field ${field['name']}: " + err.Error())
  }
% endfor
  return
}
</%def>

<%def name="method_writer(struct_name, fields, cls_index, method_index)">
<% bits = 0 %>
func (f *${struct_name}) Write(writer io.Writer) (err error) {
  var clsIndex uint32 = ${cls_index}
  if err = WriteShort(writer, &clsIndex); err != nil {
    return err
  }
  var methodIndex uint32 = ${method_index}
  if err = WriteShort(writer, &methodIndex); err != nil {
    return err
  }

  % for field in fields:
    % if field['type'].go_type == 'bool':
      % if bits == 0:
  var bits byte
      % endif
  if f.${field['name']} {
    bits |= 1 << ${bits}
  }
      <% bits += 1 %>
    % else:
      % if bits > 0:
        ${write_bits()}
        <% bits = 0 %>
      % endif
  err = Write${field['serializer']}(writer, f.${field['name']})
  if err != nil {
    return errors.New("Error writing field ${field['name']}")
  }
  % endif
  % endfor
  % if bits > 0:
    ${write_bits()}
  % endif
  return
}
</%def>

<%def name="write_bits()">
  err = WriteOctet(writer, bits)
  if err != nil {
    return errors.New("Error writing bit fields")
  }
</%def>


