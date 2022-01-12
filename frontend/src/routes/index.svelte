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

  function service() {
    fetch('/nerthus/service/'+scope+'/'+server_name, {
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

  function getLoadbalancers() {
    fetch('/nerthus/loadbalancers/', {
      method: 'GET',
      mode: 'cors',
      cache: 'no-cache',
      credentials: 'omit',
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
      loadbalancers = data.loadbalancers;
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

  let showConsentForm = false;
  export let registered = false;
  let gdprconsent = true;
  let disabled = false;
  let validEmail = false;
  let validUsername = false;
  let validPassword = false;
  let validConfirmPassword = false;
  let user = {
    name: "",
    password: "",
  }
  let loadbalancers = [];
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

<svelte:head>
	<title>Dashboard</title>
</svelte:head>

<div class="content flex">
  <div class="item">
    <h1>Nerthus</h1>
    <p style="text-align: center;">Simple Nerthus interface. For more information look at the <a href="https://github.com/Cantara/nerthus">github</a>.</p>
  </div>
  <div class="new_line"/>
  <div class="item">
    <Input required label="Username" bind:value={user.name}/>
  </div>
  <div class="item">
    <Input required password label="Password" bind:value={user.password}/>
  </div>
  <div class="item">
      <Button click={getLoadbalancers} bind:disabled>Get loadbalancers</Button>
  </div>
  <div class="new_line"/>
  <div class="item">
    <h2>New server</h2>
    <p>Use this to create a new server in a new scope with a new service</p>
    <form on:submit|preventDefault={() => {}}>
      <Select required label="Loadbalancer" values={loadbalancersDropdown} bind:value={loadbalancer}/>
      <Input required label="Scope" bind:value={scope}/>
      <Input required label="Server name" bind:value={server_name}/>
      <Input required number label="Port" bind:value={body.service.port}/>
      <Input required label="Path" bind:value={body.service.path}/>
      <Input required label="Artifact ID" bind:value={body.service.artifact_id}/>
      <Input required label="Health report url" bind:value={body.service.health_report_url}/>
      <Input required label="Filebeat config url" bind:value={body.service.filebeat_config_url}/>
      <Input required multiline autogrow label="Local override properties" bind:value={body.service.local_override_properties}/>
      <Input required multiline label="Semantic update service properties" bind:value={body.service.semantic_update_service_properties}/>
      <Button click={server} bind:disabled>Create</Button>
    </form>
  </div>
  <div class="item">
    <h2>New service on server</h2>
    <p>Use this to create a new service on an existing server</p>
    <form on:submit|preventDefault={() => {}}>
      <Select required label="Loadbalancer" values={loadbalancersDropdown}]} bind:value={loadbalancer}/>
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
      <Button click={service} bind:disabled>Create</Button>
    </form>
  </div>
  <div class="item">
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
  </div>
  <div class="item">
    <h2>Add server in scope</h2>
    <p>Use this to add an existing service to a server</p>
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
      <Button click={service} bind:disabled>Add</Button>
    </form>
  </div>
  <div class="new_line" style="padding-top: 1.5em;"/>
  {loadbalancers}
</div>

<style>
	h1 {
		color: var(--primary);
		font-size: 3em;
		font-weight: 350;
    margin: .5rem 0;
	}
  p {
    text-align: left;
  }
  hr {
    width: 100%;
    max-width: 100%;
    height: 0;
    max-height: 0;
    border: solid;
  	display: block;
		margin-top: .5em;
		margin-bottom: .5em;
		margin-left: auto;
		margin-right: auto;
		border-style: inset;
		border-width: 1px;
		color: rgba(0,0,0,.12);
	}
  .inline_content {
    display: flex;
    justify-content: center;
    align-content: center;
  }
  .content {
		position: relative;
		max-width: 1270px;
		margin-left: auto;
		margin-right: auto;
    padding-top: 1em;
    background: #fff;
  }
  .flex {
    display: flex;
    flex-flow: row wrap;
    justify-content: space-around;
    align-items: flex-start;
    align-content: space-around;
  }

  .item {
    flex: 0 0 45%;
  }
  .min_item {
    flex: 0 0 20%;
  }
  .large_item {
    flex: 0 0 100%;
    width: 100%;
  }
  .new_line {
    flex: 0 0 100%;
  }
  .item_org {
    flex-basis: auto;
    flex-grow: 1;
    flex-shrink: 1;
  }
  .center {
    align-self: center;
  }
  .data {
    text-align:center;
    padding:4px .5em;
  }
</style>

