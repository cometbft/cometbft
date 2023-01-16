<template>
  <div>
    <div class="search-box">
      <div class="search-box__icon">
        <icon-search :stroke="value ? '#66A1FF' : '#aaa'" :fill="value ? '#66A1FF' : '#aaa'"></icon-search>
      </div>
      <div class="search-box__input">
        <input
          class="search-box__input__input"
          type="text"
          autocomplete="off"
          placeholder="Search"
          ref="search"
          :value="value"
          @keydown.38.prevent="$emit('keypress', $event)"
          @keydown.40.prevent="$emit('keypress', $event)"
          @keydown.13.prevent="$emit('keypress', $event)"
          @input="$emit('input', $event.target.value)" />
      </div>
      <div class="search-box__clear">
        <icon-circle-cross class="search-box__clear__icon" v-if="value && value.length > 0" @click.native="$emit('input', '')" @keydown.enter="$emit('input', '')" tabindex="0"></icon-circle-cross>
      </div>
      <a class="search-box__button" @click="$emit('cancel', true)" @keydown.enter="$emit('cancel', true)" tabindex="0">Cancel</a>
    </div>
  </div>
</template>

<style scoped>
.search-box {
  width: 100%;
  display: grid;
  grid-auto-flow: column;
  align-items: center;
  box-shadow: inset 0 -1px 0 0 rgba(176, 180, 207, 0.2);
  padding-left: 2rem;
  padding-right: 2rem;
  grid-template-columns: 1.5rem 1fr 1.25rem;
  gap: 1rem;
  box-sizing: border-box;
  font-family: var(--ds-font-family, inherit);
}
.search-box__input__input {
  border: none;
  background: none;
  outline: none;
  font-size: 1.25rem;
  width: 100%;
  padding: 1.5rem 0.5rem;
}
.search-box__input__input::-webkit-input-placeholder {
  color: rgba(14, 33, 37, 0.26);
}
.search-box__clear__icon {
  cursor: pointer;
  fill: rgba(0, 0, 0, 0.15);
  margin-top: 0.25rem;
}
.search-box__clear__icon:hover,
.search-box__clear__icon:focus {
  fill: rgba(0, 0, 0, 0.25);
  outline: none;
}
.search-box__button {
  text-transform: uppercase;
  color: var(--accent-color, black);
  font-weight: 500;
  cursor: pointer;
  height: 100%;
  display: flex;
  align-items: center;
}
</style>

<script>
import IconSearch from "./IconSearch.vue";
import IconCircleCross from "./IconCircleCross.vue";

export default {
  components: {
    IconSearch,
    IconCircleCross
  },
  props: {
    value: {
      type: String
    }
  },
  mounted() {
    if (this.$refs.search) this.$refs.search.focus();
  }
};
</script>