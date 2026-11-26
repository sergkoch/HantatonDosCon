<script lang="ts">
  import { onMount } from "svelte";

  const KEY = "cookie-consent";
  let visible = $state(false);

  onMount(() => {
    try {
      visible = localStorage.getItem(KEY) !== "accepted";
    } catch {
      visible = true;
    }
  });

  function accept() {
    try {
      localStorage.setItem(KEY, "accepted");
    } catch {
      // localStorage недоступен — просто скрываем баннер
    }
    visible = false;
  }
</script>

{#if visible}
  <div
    class="cookie-consent"
    role="region"
    aria-label="Использование файлов cookie"
  >
    <p class="cookie-consent__text">
      Мы используем файлы cookie, чтобы сайт работал корректно и запоминал ваши
      настройки (например, выбранную тему). Продолжая работу, вы соглашаетесь с
      этим.
    </p>
    <button class="button button--primary" type="button" onclick={accept}>
      Принять
    </button>
  </div>
{/if}
