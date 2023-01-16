<template>
  <div>
    <div class="results__noresults__container">
      <div class="results__noresults">
        <div class="results__noresults__icon">
          <icon-search></icon-search>
        </div>
        <div class="results__noresults__h1">No results for <strong>“{{query}}”</strong>
        </div>
        <div class="results__noresults__p">
          <span>Try queries such as <span class="results__noresults__a" @click="querySet('auth')" @keydown.enter="querySet('auth')">auth</span>, <span class="results__noresults__a" @click="querySet('slashing')" @keydown.enter="querySet('slashing')">slashing</span>, or <span class="results__noresults__a" @click="querySet('staking')" @keydown.enter="querySet('staking')">staking</span>.</span>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.results__noresults__container {
  height: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  flex-grow: 1;
  padding-top: 4rem;
  padding-bottom: 4rem;
  padding-left: 2rem;
  padding-right: 2rem;
  font-family: var(--ds-font-family, inherit);
  overflow-x: hidden;
  word-break: break-all;
}
.results__noresults {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
}
.results__noresults__icon {
  max-width: 80px;
  margin-bottom: 2rem;
  fill: #ccc;
}
.results__noresults__h1 {
  color: rgba(22, 25, 49, 0.65);
  font-size: 1.5rem;
  margin-bottom: 1rem;
  text-align: center;
}
.results__noresults__p {
  color: rgba(22, 25, 49, 0.65);
}
.results__noresults__a {
  cursor: pointer;
  color: var(--accent-color, black);
}
</style>

<script>
import IconSearch from "./IconSearch.vue";

export default {
  components: {
    IconSearch
  },
  props: {
    query: {
      type: String
    }
  },
  methods: {
    querySet(string) {
      this.$emit("query", string);
    }
  }
};
</script>