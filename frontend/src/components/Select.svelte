<script>
  import Input from "../components/Input.svelte";
  import Dropdown from "./Dropdown.svelte";

  export let required = false;
  export let valid = false;
  export let label = "";
  export let value = {
      name: "",
      extras: [""],
  };
  export let values = [
    {
      name: "",
      extras: [""],
    }
  ];
  let open = false;

$: {
  if (required && !value) {
    valid = false;
  } else if (!value.name) {
    valid = false;
  } else if (value.name == "") {
    valid = false;
  } else {
    valid = true;
  }
}
</script>

<Dropdown bind:open>
  <Input bind:required downarrow bind:label value={value.name} slot="trigger"/>
  <span>
    <ul class="droplist">
      {#each values as val}
      <li on:click={() => {value = val; open = false}}>
        <hr>
        <nobr><p>{val.name}</p></nobr>
        <ul class="horizontal">
        {#each val.extras as extra}
          <li>{extra}</li>
        {/each}
        </ul>
      </li>
      {/each}
      <hr>
   </ul>
  </span>
</Dropdown>

<style scoped>
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
  nobr > p {
    margin: 0 0 0 0;
  }
  ul.droplist {
    padding-left: 0px;
    list-style: none;
    margin: 0 0 0 0;
  }
  ul.droplist > li {
    cursor: pointer;
    margin: .25em 0 .125em 0;
  }
  ul.horizontal {
    padding-left: 0px;
    text-align: right;
    margin-left: 1.5em;
  }
  ul.horizontal > li {
    display: inline-block;
    padding: .10em .5em;
  }
</style>
