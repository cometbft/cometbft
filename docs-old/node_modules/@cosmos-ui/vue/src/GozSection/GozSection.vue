<template>
  <div class="section-content">
    <div class="container">
      <div class="content">
        <div class="icon" v-if="this.$slots['icon']">
          <div class="icon__icon">
            <slot name="icon"/>
          </div>
        </div>
        <div class="title" v-if="this.$slots['title']">
          <slot name="title"/>
        </div>
        <div class="subtitle" v-if="this.$slots['subtitle']">
          <slot name="subtitle"/>
        </div>
        <div class="notes" v-if="this.$slots['notes']">
          <slot name="notes"/>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
a {
  color: var(--primary-light);
  text-decoration: none;
}
.section-content {
  display: flex;
  flex-direction: column;
  padding: 1rem;
  align-self: center;
  color: var(--white-100, white);
  background: var(--grey-14);
  font-family: var(--ds-font-family, inherit);
}
.container {
  display: flex;
  flex-direction: column;
  padding: 0rem;
  max-width: 44rem;
  width: 100%;
  align-self: center;
}
.icon {
  stroke: var(--secondary);
  width: 100%;
  display: flex;
  width: 100%;
  justify-content: center;
}
.icon__icon {
  max-width: 4rem;
  max-height: 4rem;
}
.content {
  align-self: center;
  margin: 0rem 6rem;
  display: contents;
}
.title {
  font-weight: 500;
  font-size: 2.5rem;
  line-height: 1.3;
  color: var(--white-100, white);
  text-align: center;
  letter-spacing: -0.04em;
  margin: 2rem 0;
}
.subtitle {
  font-size: 1.25rem;
  line-height: 1.4;
  color: var(--white-80);
}
.notes {
  font-size: 1.25rem;
  line-height: 1.75rem;
  color: var(--white-80);
}
ul {
  list-style-type: disc;
  padding-left: 1rem;
}
ul li {
  margin-bottom: 1.25rem;
}
ul ul {
  list-style-type: none;
  padding: 0;
}
ul ul li {
  margin: 1.25rem auto 1.25rem auto;
}
ul li.note {
  font-size: 0.875rem;
  line-height: 1.25rem;
  letter-spacing: 0.01em;
}
@media screen and (max-width: 600px) {
  .title {
    font-size: 1.2rem;
    margin: 1rem 0;
  }
  .subtitle {
    font-size: 1rem;
  }
  .notes {
    font-size: 1rem;
  }
}
</style>