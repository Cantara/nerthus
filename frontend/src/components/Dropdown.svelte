<script>
  import { fade, fly } from 'svelte/transition';

  export let open = false;
</script>

<div class="trigger" on:click={() => {open = !open}}>
  <slot name="trigger"/>
</div>
{#if open}
<div class="dropdown" in:fly={{x: -500, duration: 2000}} out:fade={{duration: 250}}>
  <slot />
</div>
{/if}

<style scope>
  .dropdown-base {
    position: relative;
    display: inline-block;
    width: 100%;
  }
  .trigger {
    cursor: pointer;
  }
  .dropdown {
    position: absolute;
    background-color: var(--gray-lighter);
    border: 1px solid #474a4d;
    border-radius: 8px;
    padding: 1rem;
    overflow: hidden;
    box-shadow: 0px 8px 16px 0px rgba(0,0,0,0.2);
  }
</style>
