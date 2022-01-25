<script>
  import Box from "../components/Box.svelte";
  import Button from "../components/Button.svelte";
  import Input from "../components/Input.svelte";
  import Select from "../components/Select.svelte";

  function putService() {
    fetch('/nerthus/service/'+scope+'/'+server_name+'/'+body.artifact_id, {
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

  let valid_key = false;
  let valid_scope = false;
  let valid_server = false;
  let valid_port = false;
  let valid_path = false;
  let valid_artifact = false;
  let valid_health = false;
  let valid_filebeat = false;
  let valid_local = false;
  let valid_semantic = false;

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

$: disabled = !(valid_scope && valid_server && valid_key && valid_port && valid_path && valid_artifact && valid_health && valid_filebeat && valid_local && valid_semantic)
</script>

<h2>Add service to server</h2>
<p>Use this to add a service to an excisting server</p>
<form on:submit|preventDefault={() => {}}>
  <Input required label="Scope" bind:value={scope} bind:valid={valid_scope}/>
  <Input required label="Server name" bind:value={server_name} bind:valid={valid_server}/>
  <Input required number label="Port" bind:value={body.service.port} bind:valid={valid_port}/>
  <Input required label="Path" bind:value={body.service.path} bind:valid={valid_path}/>
  <Input required label="Artifact ID" bind:value={body.service.artifact_id} bind:valid={valid_artifact}/>
  <Input required label="Health report url" bind:value={body.service.health_report_url} bind:valid={valid_health}/>
  <Input required label="Filebeat config url" bind:value={body.service.filebeat_config_url} bind:valid={valid_filebeat}/>
  <Input required multiline autogrow label="Local override properties" bind:value={body.service.local_override_properties} bind:valid={valid_local}/>
  <Input required multiline label="Semantic update service properties" bind:value={body.service.semantic_update_service_properties} bind:valid={valid_semantic}/>
  <Input required multiline autogrow label="key" bind:value={body.key} bind:valid={valid_key}/>
  <Button click={putService} bind:disabled>Add</Button>
</form>
