<script lang="ts">
  import { navigate } from "../lib/router.svelte";
</script>

<div class="container">
  <section class="section">
    <h1 class="section__title">Страница не найдена</h1>
    <p>Проверьте адрес или вернитесь на главную.</p>
    <a
      class="button button--primary"
      href="/"
      onclick={(event) => {
        event.preventDefault();
        navigate("/");
      }}
    >
      На главную
    </a>
  </section>
</div>
