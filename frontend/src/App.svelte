<script>
  import Index from "./routes/index.svelte";
</script>

<main>
  <Index/>
</main>

<style>
:root {
  /* https://www.canva.com/colors/color-wheel/ */
  --orange: #FCA611;
  --green: #11FC31;
  --blue: #1167FC;
  --purple: #FC11DD;
  --primary: var(--blue);
  --secoundary: var(--purple);
  --monochromatic: #D78903;
  --triadicOne: #11FCA6;
  --triadicTwo: #A611FC;
  --analogOne: #DDFC11;
  --analogTwo: #FC3111;
  --gray-lightest: #C5C6D0;
  --gray-lighter: #E6E6E6;
  --gray: #989799;
  --gray-2: #41424C;
  --gray-darker: #373737;
  --gray-darkest: #232023;
  --background: var(--gray-lighter);
}
	main {
		text-align: center;
		padding: 0;
		margin: 0 auto;
    min-height: 100%;
    background-color: var(--background);
    /* background-image: linear-gradient(var(--gray),#fff); */
	}
	@media (min-width: 640px) {
    main {
			max-width: none;
		}
	}
</style>
