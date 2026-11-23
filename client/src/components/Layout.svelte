<script lang="ts">
  import { navigate } from "../lib/router.svelte";
</script>

<div class="site">
  <a class="skip-link" href="#main">К содержимому</a>

  <header class="site-header">
    <div class="container site-header__inner">
      <a
        class="brand"
        href="/"
        onclick={(event) => {
          event.preventDefault();
          navigate("/");
        }}
      >
        Доступный контент
      </a>

      <nav class="site-nav" aria-label="Основная навигация">
        <a
          href="/"
          onclick={(event) => {
            event.preventDefault();
            navigate("/");
          }}
        >
          Проверить
        </a>
        <a href="#how">Как это работает</a>
      </nav>
    </div>
  </header>

  <main class="site-main" id="main">
    <slot />
  </main>

  <footer class="site-footer">
    <div class="container">
      <p>Учебный проект команды. Проверка доступности контента.</p>
    </div>
  </footer>
</div>
