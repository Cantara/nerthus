<script>
  import Box from "../components/Box.svelte";
  import Button from "../components/Button.svelte";
  import Input from "../components/Input.svelte";
  import Select from "../components/Select.svelte";

  function putDatabase() {
    fetch('/nerthus/database/'+scope+'/'+artifact_id, {
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
  let artifact_id = "";

  let valid_key = false;
  let valid_scope = false;
  let valid_artifact_id = false;

  let disabled = false;

  export let user = {
    name: "",
    password: "",
  }

$: disabled = !(valid_scope && valid_artifact_id && valid_key)
</script>

<h2>Add server to scope</h2>
<p>Use this to create a new server in an existing scope</p>
<form on:submit|preventDefault={() => {}}>
  <Input required label="Scope" bind:value={scope} bind:valid={valid_scope}/>
  <Input required label="Artifact ID" bind:value={artifact_id} bind:valid={valid_artifact_id}/>
  <Input required multiline autogrow label="key" bind:value={body.key} bind:valid={valid_key}/>
  <Button click={putDatabase} bind:disabled>Add</Button>
</form>
