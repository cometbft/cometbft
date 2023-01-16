<template>
  <div class="page-header">
    <div class="page-header__container">
      <div class="page-header__text">
        <div class="page-header__suptitle">
          <template v-if="suptitle">{{ suptitle }}</template>
          <slot v-else name="suptitle"></slot>
        </div>
        <div class="page-header__project-title">
          <template v-if="projectTitle">{{ projectTitle }}</template>
          <slot v-else name="project-title"></slot>
        </div>
        <div class="page-header__title">
          <template v-if="title">{{ title }}</template>
          <slot v-else name="title"></slot>
        </div>
        <div class="page-header__subtitle">
          <template v-if="subtitle">{{ subtitle }}</template>
          <slot v-else name="subtitle"></slot>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  props: ["suptitle", "title", "subtitle", "project-title"]
};
</script>

<style scoped>
.page-header {
  background: hsl(220, 33%, 98%);
}

.page-header__container {
  padding: 2rem;
}

.page-header__text {
  margin-top: 1.5rem;
}

.page-header__title {
  margin-bottom: 1rem;
}

.page-header__title {
  font-size: 2.5rem;
  line-height: 1.16;
  font-weight: 300;
}

.page-header__subtitle {
  color: hsla(0, 100%, 0%, 0.667);
}

.page-header__project-title {
  font-family: var(--ds-font-family, inherit);
  font-weight: 500;
  font-size: var(--ds-h4-font-size);
  line-height: 1.25rem;
  letter-spacing: 0.25rem;
  text-transform: uppercase;
  margin-bottom: 1.25rem;
  color: hsla(0, 100%, 0%, 0.8);
}

@media screen and (max-width: 1680px) {
  .page-header__container {
    padding: 5.75rem 8rem 2rem;
  }
}

@media screen and (max-width: 1024px) {
  .page-header__container {
    padding: 5.75rem 5.75rem 2rem;
  }
}

@media screen and (max-width: 768px) {
  .page-header__container {
    padding: 2rem;
  }
}
</style>
