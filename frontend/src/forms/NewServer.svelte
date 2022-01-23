<script>
  import Box from "../components/Box.svelte";
  import Button from "../components/Button.svelte";
  import Input from "../components/Input.svelte";
  import Select from "../components/Select.svelte";

  function server() {
    fetch('/nerthus/server/'+scope+'/'+server_name, {
      method: 'PUT',
      mode: 'cors',
      cache: 'no-cache',
      credentials: 'omit',
      body: JSON.stringify(body),
      headers: {
        'Authorization': 'Basic ' + btoa(user.name + ":" + user.password),
        'Accept': 'application/json',
        'Content-Type': 'application/json',
      },
    })
    .then(response => response.json())
    .then(data => {
      console.log(data)
      if (data.error) {
        console.log(data.error);
        //message.error(data.error);
        return;
      }
    })
    .catch((error) => {
      console.log(error);
      //message.error(error);
    });
  }

  let body = {
    service: {
      elb_listener_arn: "",
      elb_securitygroup_id: "",
      port: 0,
      path: "",
      artifact_id: "",
      health_report_url: "",
      filebeat_config_url: "",
      local_override_properties: "",
      semantic_update_service_properties: "",
    },
    key: "",
  }
  let scope = "";
  let server_name = "";

  let disabled = false;

  export let user = {
    name: "",
    password: "",
  }
  export let loadbalancers = [];
  let loadbalancer = {};
  let loadbalancersDropdown = [];

$: {
  loadbalancersDropdown = loadbalancers.map(value => ({
    name: value.dns_name,
    extras: value.paths,
    arn: value.arn,
    listener_arn: value.listener_arn,
    security_group: value.security_group
  }))
}
$: body.service.elb_listener_arn = loadbalancer.listener_arn
$: body.service.elb_securitygroup_id = loadbalancer.security_group
</script>

<h2>Add server to scope</h2>
<p>Use this to create a new server in an existing scope with existing service</p>
<form on:submit|preventDefault={() => {}}>
  <Input required label="Scope" bind:value={scope}/>
  <Input required label="Server name" bind:value={server_name}/>
  <Input required number label="Port" bind:value={body.service.port}/>
  <Input required label="Path" bind:value={body.service.path}/>
  <Input required label="Artifact ID" bind:value={body.service.artifact_id}/>
  <Input required label="Health report url" bind:value={body.service.health_report_url}/>
  <Input required label="Filebeat config url" bind:value={body.service.filebeat_config_url}/>
  <Input required multiline autogrow label="Local override properties" bind:value={body.service.local_override_properties}/>
  <Input required multiline label="Semantic update service properties" bind:value={body.service.semantic_update_service_properties}/>
  <Input required multiline label="key" bind:value={body.key}/>
  <Button click={server} bind:disabled>Add</Button>
</form>
