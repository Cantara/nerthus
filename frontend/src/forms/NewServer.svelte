<script>
  import Box from "../components/Box.svelte";
  import Button from "../components/Button.svelte";
  import Input from "../components/Input.svelte";
  import Select from "../components/Select.svelte";

  function putServer() {
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
    key: "",
  }
  let scope = "";
  let server_name = "";

  let valid_key = false;
  let valid_scope = false;
  let valid_server = false;

  let disabled = false;

  export let user = {
    name: "",
    password: "",
  }

$: disabled = !(valid_scope && valid_server && valid_key)
</script>

<h2>Add server to scope</h2>
<p>Use this to create a new server in an existing scope</p>
<form on:submit|preventDefault={() => {}}>
  <Input required label="Scope" bind:value={scope} bind:valid={valid_scope}/>
  <Input required label="Server name" bind:value={server_name} bind:valid={valid_server}/>
  <Input required multiline autogrow label="key" bind:value={body.key} bind:valid={valid_key}/>
  <Button click={putServer} bind:disabled>Add</Button>
</form>
