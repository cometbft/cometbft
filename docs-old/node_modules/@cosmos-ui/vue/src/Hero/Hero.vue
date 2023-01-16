<template>
  <div :class="['component__hero__container', `component__hero__container__alt__${alt}`]" :style="{'background-image': `url(${value.background})`}">
    <div class="wrapper">
      <div class="text">
        <div>
          <div class="text__h2">{{value.h2}}</div>
          <div class="text__h1">{{value.h1}}</div>
          <div class="text__h3">{{value.h3}}</div>
          <div class="text__buttons">
            <a class="text__buttons__item" :href="item.href" v-for="item in value.buttons" :key="item.title">
              {{item.title}}
            </a>
          </div>
        </div>
      </div>
      <div class="image">
        <div class="image__image">
          <img class="image__image__img" :src="value.img" alt="">
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
a {
  color: inherit;
  text-decoration: none;
}

.component__hero__container {
  font-family: var(--ds-font-family, inherit);
  background-size: cover;
  color: var(--black, black);
}

.wrapper {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(400px, 1fr));
  padding-top: 2rem;
  padding-bottom: 2rem;
}

.component__hero__container.component__hero__container__alt__true {
  color: var(--white-100, white);
}

.text {
  display: grid;
  align-items: center;
}

.text__h1 {
  margin-top: 1rem;
  margin-bottom: 2rem;
  font-size: var(--ds-h1-font-size, 3rem);
}

.text__buttons__item {
  display: inline-block;
  margin-right: 1rem;
  padding: 1rem 2rem;
  text-transform: uppercase;
  background: rgba(0, 0, 0, 0.4);
  letter-spacing: 0.05em;
  box-shadow: inset 0 0 0 2px white;
  border-radius: 0.25rem;
}

.image {
  display: grid;
  justify-content: flex-end;
  align-items: center;
}

.image__image {
  width: 500px;
  position: relative;
}

.image__image__img {
  padding: 50px;
  box-sizing: border-box;
  width: 100%;
}

@media screen and (max-width: 600px) {
  .wrapper {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  }

  .image {
    justify-content: center;
  }

  .image__image {
    width: 300px;
  }

  .text__h1 {
    font-size: 2rem;
  }
}
</style>

<script>
export default {
  props: ["value", "alt"]
};
</script>