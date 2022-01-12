<script>
  export let click = () => {};
  export let flat = false;
  export let primary = "var(--primary-text)";
  export let secondary = "var(--primary)";
  export let disabled = false;
</script>

<button disabled={disabled} class="{flat ? 'flat' : 'elevated'}" style="--primary-btn: {primary}; --secondary-btn: {secondary}" on:click={click}>
  <slot/>
</button>

<style>
  button {
    height: 100%;
    cursor: pointer;

    padding: .5em 1em;
    margin: .25em;
		font-weight: 500;
		letter-spacing: .0892857143em;
		text-indent: .0892857143em;
		text-transform: uppercase;
		white-space: nowrap;

    border-radius: .25em;
    border-style: none;
  }
  button:disabled {
    background-color: #A4AEB8;
    cursor: auto;
    opacity: .75;
  }
  button:focus {
    outline: none;
  }

  .elevated {
    background-color: var(--secondary-btn);
    color: var(--primary-btn);
    box-shadow: 0 3px 1px -2px rgba(0,0,0,.2),0 2px 2px 0 rgba(0,0,0,.14),0 1px 5px 0 rgba(0,0,0,.12);
  }

  .elevated:hover:enabled {
    opacity: .75;
  }

  .flat {
    background-color: transparent;
    color: var(--secondary-btn);
  }
  .flat:hover:enabled {
    background-color: rgba(var(--primary-btn), .12); /* TODO: fix this static bs */
  }
</style>
