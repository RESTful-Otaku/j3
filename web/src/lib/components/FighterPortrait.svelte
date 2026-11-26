<script lang="ts">
  export let src: string;
  export let label: string;
  export let size: string = '44px';
  export let fallback: string = '';
  export let clubColour: string = '#64748b';

  $: hasPhoto = src && src.trim() !== '';
</script>

<div class="portrait-wrapper" style="--portrait-size: {size};">
  {#if hasPhoto}
    <img
      class="portrait-photo"
      src={src}
      alt={label}
      width={size}
      height={size}
      on:error={(e) => {
        // Fallback handled by CSS if image fails
      }}
    />
  {:else}
    <span
      class="portrait-fallback"
      style="background: {clubColour}; color: {fallback ? '#ffffff' : '#0f172a'};"
    >
      {fallback || label.split(' ').map((p) => p[0]).join('')}
    </span>
  {/if}
</div>

<style>
  .portrait-wrapper {
    display: grid;
    place-items: center;
    width: var(--portrait-size);
    height: var(--portrait-size);
    border-radius: 50%;
    overflow: hidden;
    border: 2px solid rgba(255, 255, 255, 0.3);
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.2);
    flex: none;
  }
  .portrait-photo {
    width: 100%;
    height: 100%;
    object-fit: cover;
    display: block;
  }
  .portrait-fallback {
    width: 100%;
    height: 100%;
    display: grid;
    place-items: center;
    font-weight: 800;
    font-size: 0.75rem;
    line-height: 1;
  }
</style>
