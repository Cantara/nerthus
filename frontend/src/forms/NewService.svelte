<script>
  import Box from "../components/Box.svelte";
  import Button from "../components/Button.svelte";
  import Input from "../components/Input.svelte";
  import Select from "../components/Select.svelte";

  function putService() {
    let bodyT = body
    bodyT.service.health.service_type = bodyT.service.health.service_type.name
    fetch('/nerthus/service/'+scope+'/'+server_name+'/'+body.service.artifact_id, {
      method: 'PUT',
      mode: 'cors',
      cache: 'no-cache',
      credentials: 'omit',
      body: JSON.stringify(bodyT),
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
      health: {
        service_name: "",
        service_tag: "",
        service_type: {
          name: "",
        },
      },
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
  let valid_loadbalancer = false;
  let valid_port = false;
  let valid_path = false;
  let valid_artifact = false;
  let valid_health_name = false;
  let valid_health_tag = false;
  let valid_health_type = false;
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
  let health_service_types = [
    {
      name: "A2A",
      extras: [""],
    },
    {
      name: "H2A",
      extras: [""],
    },
    {
      name: "ACS",
      extras: [""],
    },
    {
      name: "CS",
      extras: [""],
    }
  ];

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

$: disabled = !(valid_scope && valid_server && (valid_loadbalancer || loadbalancer != {}) && valid_key && valid_port && valid_path && valid_artifact && valid_health_name && valid_health_tag && (valid_health_type || body.service.health.service_type.name != "") && valid_local && valid_semantic && valid_key)
</script>

<h2>Add service to server</h2>
<p>Use this to add a service to an excisting server</p>
<form on:submit|preventDefault={() => {}}>
  <Input required label="Scope" bind:value={scope} bind:valid={valid_scope}/>
  <Input required label="Server name" bind:value={server_name} bind:valid={valid_server}/>
  <Select required label="Loadbalancer" values={loadbalancersDropdown} bind:value={loadbalancer} bind:valid={valid_loadbalancer}/>
  <Input required number label="Port" bind:value={body.service.port} bind:valid={valid_port}/>
  <Input required label="Path" bind:value={body.service.path} bind:valid={valid_path}/>
  <Input required label="Artifact ID" bind:value={body.service.artifact_id} bind:valid={valid_artifact}/>
  <Input required label="Health service name" bind:value={body.service.health.service_name} bind:valid={valid_health_name}/>
  <Input required label="Health service tag" bind:value={body.service.health.service_tag} bind:valid={valid_health_tag}/>
  <Select required label="Health service type" values={health_service_types} bind:value={body.service.health.service_type} bind:valid={valid_health_type}/>
  <Input required multiline autogrow label="Local override properties" bind:value={body.service.local_override_properties} bind:valid={valid_local}/>
  <Input required multiline label="Semantic update service properties" bind:value={body.service.semantic_update_service_properties} bind:valid={valid_semantic}/>
  <Input required multiline autogrow label="key" bind:value={body.key} bind:valid={valid_key}/>
  <Button click={putService} bind:disabled>Add</Button>
</form>
