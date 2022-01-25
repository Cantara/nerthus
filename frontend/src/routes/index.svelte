<script>
  import Box from "../components/Box.svelte";
  import Button from "../components/Button.svelte";
  import Input from "../components/Input.svelte";
  import Select from "../components/Select.svelte";
  import NewScope from "../forms/NewScope.svelte";
  import NewServer from "../forms/NewServer.svelte";
  import NewService from "../forms/NewService.svelte";

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
    <NewScope bind:user />
  </div>
  <div class="item">
    <NewServer bind:user />
  </div>
  <div class="large_item">
    <NewService bind:user bind:loadbalancers />
  </div>
  <div class="new_line" style="padding-top: 1.5em;"/>
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

