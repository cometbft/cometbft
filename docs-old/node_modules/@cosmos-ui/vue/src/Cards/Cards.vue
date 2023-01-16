<template>
  <div class="component__cards__container">
    <slot/>
    <div :class="['cards', `cards__format__${format}`]">
      <a :href="item.href" :class="['cards__item', format && `cards__item__format__${format}`]" v-for="(item, i) in value" :key="item.href + i">
        <div class="cards__item__image">
          <img v-if="item.img" :src="item.img" alt="" class="cards__item__image__img">
          <img v-if="item.icon" class="cards__item__image__icon" :src="item.icon">
          <div v-if="item.alt" class="cards__item__image__alt">{{item.alt}}</div>
        </div>
        <div>
          <div class="cards__item__h1">{{item.h1}}</div>
          <div class="cards__item__h2">{{item.h2}}</div>
        </div>
      </a>
    </div>
  </div>
</template>

<style scoped>
a {
  color: inherit;
  text-decoration: none;
}

.component__cards__container {
  font-family: var(--ds-font-family, inherit);
  color: var(--black, black);
}

.cards {
  display: grid;
  gap: 2rem;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
}

.cards.cards__format__horizontal {
  width: 100vw;
  grid-template-columns: repeat(auto-fit, minmax(500px, 1fr));
}

.cards__item {
  position: relative;
  display: block;
}

.cards__item.cards__item__format__horizontal {
  display: grid;
  gap: 1rem;
  grid-template-columns: 1.5fr 1fr;
}

.cards__item__h1 {
  margin-bottom: 0.5rem;
  color: black;
  font-size: var(--ds-h4-font-size, 1rem);
  font-weight: var(--ds-h4-font-weight, 600);
}

.cards__item__h2 {
  font-size: var(--ds-h5-font-size, 1rem);
  font-weight: var(--ds-h5-font-weight, rgba(0, 0, 0, 0.5));
}

.cards__item__image {
  background-color: black;
  height: 25vh;
  min-width: 15vw;
  min-height: 200px;
  margin-bottom: 0.5rem;
  border-radius: 0.25rem;
  display: grid;
  align-items: center;
  justify-items: center;
}

.cards__item__image__img {
  border-radius: 0.25rem;
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
  grid-area: 1/1/1/1;
}

.cards__item__image__alt {
  width: 100%;
  height: 100%;
  display: grid;
  align-items: center;
  justify-items: center;
  background-color: rgba(0, 0, 0, 0.2);
  color: var(--white-100, white);
  grid-area: 1/1/1/1;
  font-size: var(--ds-h2-font-size, 2rem);
}

.cards__item__image__icon {
  width: 100%;
  height: 100%;
  display: grid;
  align-items: center;
  justify-items: center;
  grid-area: 1/1/1/1;
}
</style>

<script>
export default {
  props: ["value", "format"]
};
</script>
