<script>
  import Box from "../components/Box.svelte";
  import Button from "../components/Button.svelte";
  import Input from "../components/Input.svelte";
  import Select from "../components/Select.svelte";

  function putScope() {
    fetch('/nerthus/scope/'+scope, {
      method: 'PUT',
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
    })
    .catch((error) => {
      console.log(error);
      //message.error(error);
    });
  }

  let scope = "";

  let valid_scope = false;

  let disabled = false;

  export let user = {
    name: "",
    password: "",
  }

$: disabled = !valid_scope
</script>

<h2>New scope</h2>
<p>Use this to create a new scope to add servers to</p>
<form on:submit|preventDefault={() => {}}>
  <Input required label="Scope" bind:value={scope} bind:valid={valid_scope}/>
  <Button click={putScope} bind:disabled>Create</Button>
</form>
