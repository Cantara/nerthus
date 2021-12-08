curl --header "Content-Type: application/json" \
  --request PUT \
  --data '{"port":18080,"path":"information","elb_listener_arn":"arn:aws:elasticloadbalancing:us-west-2:493376950721:listener/app/devtest-events2-lb/a3807cba101b280b/90abaa841820e9b2","elb_securitygroup_id":"sg-1325864d"}' \
  -u sindre:pass \
  http://localhost:3030/nerthus/server/dev-entraos-info-newer/
