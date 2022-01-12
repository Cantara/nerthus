<script>

</script>

<div id="box">
  <slot/>
</div>

<style>
#box {
	display: block;
	max-width: 100%;
	outline: none;
	text-decoration: none;
	transition-property: box-shadow,opacity;
	overflow-wrap: break-word;
	position: relative;
	white-space: normal;
	transition: box-shadow .28s cubic-bezier(.4,0,.2,1);
	will-change: box-shadow;
	box-shadow: 0 3px 1px -2px rgba(0,0,0,.2),0 2px 2px 0 rgba(0,0,0,.14),0 1px 5px 0 rgba(0,0,0,.12);
  margin: 1em;
  padding: 1.5em;
  padding-top: .5em;
  padding-bottom: .5em;
}

</style>
