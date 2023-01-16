<template>
  <div>
    <div
      :class="[`results__item`, `results__item__selected__${index === selected}`]"
      ref="result"
      :key="index"
      v-for="(item, index) in value"
      @click="$emit('activate', {...item})"
    >
      <div class="results__item__title" v-if="item.title" v-html="item.title"></div>
      <div class="results__item__desc" v-if="item.url" v-html="formatURL(item.url)"></div>
      <div class="results__item__desc" v-if="item.desc" v-html="item.desc"></div>
    </div>
  </div>
</template>

<style scoped>
.results__item {
  padding: 1rem 2rem;
  cursor: pointer;
}
.results__item__selected__true {
  background-color: #fff;
}
.results__item__title {
  color: var(--accent-color, black);
}
.results__item__h2 {
  margin-top: 0.25rem;
  margin-bottom: 0.25rem;
  font-weight: 500;
  font-size: 0.875rem;
}
.results__item__h2__item {
  display: inline-block;
}
.results__item__h2__item:after {
  content: ">";
  margin-left: 0.25rem;
  margin-right: 0.25rem;
}
.results__item__h2__item:last-child:after {
  content: "";
}
.results__item__desc {
  opacity: 0.5;
  white-space: nowrap;
  overflow: hidden;
  position: relative;
  font-size: 0.875rem;
}
.results__item__desc:after {
  content: "";
  background: linear-gradient(to right, rgba(248, 249, 252, 0.5) 0%, #f8f9fc);
  height: 1em;
  width: 2em;
  padding-bottom: 0.25rem;
  text-align: right;
  position: absolute;
  top: 0;
  right: 0;
}
</style>

<script>
export default {
  props: {
    value: {
      type: Array,
      default: () => []
    },
    selected: {
      type: Number
    },
    base: {
      default: ""
    }
  },
  methods: {
    formatURL(url) {
      return new URL(url).pathname
        .replace(this.base, "")
        .split("/")
        .filter(e => e)
        .join(" › ")
        .replace(".html", "");
    }
  }
};
</script>