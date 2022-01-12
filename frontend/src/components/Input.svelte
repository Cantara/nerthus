<script>
  export let min_height = "4.4em";
  export let multiline = false;
  export let label = ""
  export let value;
  export let password = false;
  export let required = false;
  export let email = false;
  export let min = 0;
  export let max = -1;
  export let valid = false;
  export let autogrow = false;
  export let downarrow = false;
  export let number = false;
  let visible = false;

  const autoGrow = (event) => {
    if (!autogrow) {
      return
    }
    if (event.target.scrollHeight > event.target.clientHeight) {
      event.target.style.height = event.target.scrollHeight + "px";
    } else if (event.keyCode == 8) {
      event.target.style.height = min_height; // Prevent height from growing when deleting lines.
      if (event.target.scrollHeight > (parseFloat(getComputedStyle(event.target).fontSize) * event.target.style.minHeight.slice(0, -2))) {
        event.target.style.height = event.target.scrollHeight + 'px';
      }
    }
  }

$: {
  if (required && !value) {
    valid = false;
  } else if (value.length < min) {
    valid = false;
  } else if (max !== -1 && value.length > max) {
    valid = false;
  } else if (email && value !== null && value.length !== 0 && !(/.+@.+\..+/.test(value))) {
    valid = false;
  } else if (number && value !== null && value.length !== 0 && (/[^0-9]/.test(value))) {
    valid = false;
  } else {
    valid = true;
    if (number) {
      value = parseInt(value)
    }
  }
}
</script>
<svelte:head>
  <script src="https://kit.fontawesome.com/673f197acb.js" crossorigin="anonymous"></script>
</svelte:head>

{#if multiline}
  {#if label != ""}

  <fieldset class:invalid={!valid}>
    <legend>{(required ? '* ' : '') + label}</legend>
    {#if password}
    {:else}
    <textarea contenteditable bind:value={value} style="min-height: {min_height}" class:autogrow on:keyup={autoGrow}/>
    {/if}
  </fieldset>
  {:else}
  <div>
    <textarea contenteditable bind:value={value} style="min-height: {min_height}" class:autogrow on:keyup={autoGrow}/>
  </div>
  {/if}
{:else}
  {#if label != ""}
  <fieldset class:invalid={!valid}>
    {#if downarrow}
      <legend><i class='fas fa-angle-down' style="font-size:1.25em"/>{(required ? ' * ' : '') + label}</legend>
    {:else}
      <legend>{(required ? '* ' : '') + label}</legend>
    {/if}
    {#if password}
      <input contenteditable bind:value={value} type="password"/>
    {:else}
      <input contenteditable bind:value={value}/>
    {/if}
  </fieldset>
  {:else}
  {/if}
{/if}

<style>
div {
	border-style: solid !important;
	border-width: thin;
	border-color: gray;
	border-radius: 4px;
	height: 100%;
	overflow: auto;
  min-height: 1.3em;
}

.autogrow {
  overflow: hidden;
  overflow: -moz-hidden-unscrollable;
}

input, textarea{
  width: calc(100% - 1em);
  background-color: transparent;
  border-style: none !important;
  margin: 0;
  resize: none;
  border: none;
	font-family: inherit;
	font-size: inherit;
	padding: 0.4em 1em;
  background-image:none;
  background-color:transparent;
  -webkit-box-shadow: none;
  -moz-box-shadow: none;
  box-shadow: none;
}

input:focus, textarea:focus {
  outline:none;
}

fieldset, div {
  text-align: left;
}

fieldset {
  margin-top: 1em;
  margin-bottom: .5em;
  border: 1px solid #ccc !important;
  border-style: solid;
  border-radius: .5em;
  padding: 0.01em 1em 0.01em 0em;
}
legend {
  white-space: nowrap;
  width: auto;
  margin-left: .75em;
}

input:disabled {
	color: #ccc;
}
.invalid {
  border-color: var(--error) !important;
}
.fas {
  font-family: "Font Awesome 5 Free";
}
</style>
