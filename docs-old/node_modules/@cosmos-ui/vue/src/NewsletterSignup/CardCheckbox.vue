<template>
  <div>
    <div tabindex="0" @keypress.enter="$emit('input', !value)" :class="['container', `value__${value}`]" @click="$emit('input', !value)">
      <div class="icon">
        <div class="icon__image">
          <slot name="icon"/>
        </div>
      </div>
      <div class="body">
        <div>
          <div class="h1" v-if="$slots.h1">
            <slot name="h1"/>
          </div>
          <div class="p">
            <slot/>
          </div>
        </div>
      </div>
      <div class="checkbox">
        <div class="checkbox__icon">
          <icon-check v-if="value"/>
          <icon-uncheck v-else/>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.container {
  background-color: rgba(255, 255, 255, 0.1);
  backdrop-filter: blur(5px);
  border-radius: 8px;
  display: grid;
  grid-template-columns: 25% 60% 15%;
  min-height: 7.25rem;
  user-select: none;
  cursor: pointer;
  transition: background-color 0.25s, border-color 0.25s,
    box-shadow 0.15s ease-out, transform 0.15s ease-out;
  padding-top: 1rem;
  padding-bottom: 1rem;
  box-sizing: border-box;
  outline: none;
}
::v-deep .container:hover,
::v-deep .container:focus {
  box-shadow: 0px 12px 24px rgba(0, 0, 0, 0.07), 0px 4px 8px rgba(0, 0, 0, 0.05),
    0px 1px 0px rgba(0, 0, 0, 0.05);
}
::v-deep .container:hover {
  transform: translateY(-2px);
}
::v-deep .value__false.container:hover .checkbox__icon {
  stroke: rgba(255, 255, 255, 0.8);
}
::v-deep .value__false.container:focus:not(:hover) .checkbox__icon {
  stroke: #66a1ff;
}
::v-deep .value__true.container:focus:not(:hover) .checkbox__icon {
  fill: #66a1ff;
}
.container:active {
  transform: translateY(0);
  transition-duration: 0s;
}
.value__true.container {
  background-color: #161931;
}
.icon {
  grid-column-start: 1;
  display: flex;
  align-items: center;
  justify-content: center;
}
.icon__image {
  width: 4rem;
  height: 4rem;
  stroke: rgba(255, 255, 255, 0.5);
  fill: none;
  margin: 1rem;
}
.value__true .icon__image {
  stroke: #66a1ff;
}
::v-deep .container:active .icon__image,
::v-deep .container:active .h1,
::v-deep .container:active .p,
::v-deep .container:active .checkbox {
  opacity: 0.6;
}
.body {
  grid-column-start: 2;
  display: flex;
  align-items: center;
}
.h1 {
  font-weight: 600;
  font-size: 1.25rem;
  line-height: 1.75rem;
  letter-spacing: -0.02em;
  color: #ffffff;
  margin-bottom: 0.5rem;
}
.p {
  font-size: 0.875rem;
  line-height: 1.25rem;
  letter-spacing: 0.01em;
  color: rgba(255, 255, 255, 0.8);
}
.checkbox {
  grid-column-start: 3;
  display: flex;
  align-items: center;
  justify-content: center;
}
.checkbox__icon {
  width: 1.5rem;
  height: 1.5rem;
}
.value__true .checkbox__icon {
  fill: #ffffff;
}
.value__false .checkbox__icon {
  fill: none;
  stroke: rgba(255, 255, 255, 0.32);
}
@media screen and (max-width: 400px) {
  .h1 {
    font-size: 1rem;
  }
  .p {
    font-size: 0.8125rem;
  }
}
</style>

<script>
import IconCheck from "../Icons/IconCheck";
import IconUncheck from "../Icons/IconUncheck";

export default {
  components: {
    IconCheck,
    IconUncheck
  },
  props: {
    theme: {
      type: String,
      default: "light"
    },
    value: {
      type: Boolean,
      default: false
    }
  }
};
</script>
