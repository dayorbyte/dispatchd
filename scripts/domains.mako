
package amqp

% for domain in domains:
% if domain.custom and domain.amqp_type != 'bit':
var Read${domain.name_normalized} = Read${domain.amqp_type.capitalize()}
var Write${domain.name_normalized} = Write${domain.amqp_type.capitalize()}
% endif
% endfor