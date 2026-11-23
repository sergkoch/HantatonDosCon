<script lang="ts">
  import type { Snippet } from "svelte";

  interface ButtonProps {
    variant?: "primary" | "ghost";
    type?: "submit" | "reset" | "button";
    disabled?: boolean;
    onclick?: (event: MouseEvent) => void;
    children: Snippet;
  }

  let {
    variant = "primary",
    type = "button",
    disabled = false,
    onclick,
    children,
  }: ButtonProps = $props();
</script>

<button class={`button button--${variant}`} {type} {disabled} {onclick}>
  {@render children()}
</button>
