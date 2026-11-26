<script lang="ts">
  import { navigate } from "../lib/router.svelte";
  import { t } from "../lib/i18n.svelte";
</script>

<div class="container">
  <section class="section">
    <h1 class="section__title">{t("notFound.title")}</h1>
    <p>{t("notFound.text")}</p>
    <a
      class="button button--primary"
      href="/"
      onclick={(event) => {
        event.preventDefault();
        navigate("/");
      }}
    >
      {t("notFound.home")}
    </a>
  </section>
</div>
