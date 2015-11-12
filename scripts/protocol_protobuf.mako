<%

# Global State
ALL_METHODS = []

def field_str(field, index, nullable=False):
  go_type = field['type'].go_type
  protobuf_type = field['type'].protobuf_type
  options = []

  if go_type != protobuf_type and protobuf_type != 'bytes':
    options.append('(gogoproto.casttype) = "{}"'.format(go_type))
  if protobuf_type not in ['Table', 'bytes'] and not nullable:
    options.append('(gogoproto.nullable) = false')
  options_str = ''
  if options:
    options_str = ' [{}]'.format(', '.join(o for o in options))

  s = 'optional {} {} = {}{};'
  return s.format(protobuf_type, field['name'], index, options_str)


def normalize_field(name):
  return '_'.join([w.lower() for w in name.split('-')])

def normalize_struct(name):
  return ''.join([w.capitalize() for w in name.split('-')])


def iter_tag(node, name):
  for child in node:
    if child.tag == name:
      yield child
%>package amqp;

import "github.com/gogo/protobuf/gogoproto/gogo.proto";
import "amqp/amqp.proto";

option (gogoproto.marshaler_all) = true;
option (gogoproto.sizer_all) = true;
option (gogoproto.unmarshaler_all) = true;

% for cls_node in iter_tag(root, 'class'):
<%
  cls_name = normalize_struct(cls_node.attrib['name'])
  cls_index = cls_node.attrib['index']
%>
${class_body(cls_node, cls_name, cls_index)}
  % for method_node in iter_tag(cls_node, 'method'):
${method(method_node, cls_name, cls_index)}
  % endfor
% endfor
<%def name="class_body(cls_node, cls_name, cls_index)"><%
class_fields = list(iter_tag(cls_node, 'field'))
if not class_fields:
  return ''
%>
message ${normalize_struct(cls_name)}ContentHeaderProperties {
% for index, field in enumerate(class_fields):
<%
  domain = field.attrib.get('domain')
  if not domain:
    domain = field.attrib.get('type')
  type = domains[domain]
  field = dict(
    name=normalize_field(field.attrib['name']),
    type=type,
  )
%>\
  ${field_str(field, index+1, nullable=True)}
% endfor
}

</%def>
<%def name="comment_header(name)">
// **********************************************************************
//                    ${name}
// **********************************************************************
</%def>
<%def name="method(method_node, cls_name, cls_index)">
<%
method_name = normalize_struct(method_node.attrib['name'])
method_index = method_node.attrib['index']
struct_name = '{}{}'.format(cls_name, method_name)
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
    ))
  except:
    print field, field.attrib['name'], field.attrib
    raise

%>\
${comment_header('{} - {}'.format(cls_name, method_name))}\
${method_struct(struct_name, fields, cls_index, method_index)}
</%def>
<%def name="method_struct(struct_name, fields, cls_index, method_index)">
message ${struct_name} {
  option (gogoproto.goproto_unrecognized) = false;
  option (gogoproto.goproto_getters) = false;
  % for index, field in enumerate(fields):
  ${field_str(field, index+4)}
  % endfor
}
</%def>

