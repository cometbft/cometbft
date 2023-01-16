<template>
  <div>
    <component
      :is="tag"
      :class="['button', `button__size__${size}`, `button__type__${type}`]"
      :style="{ 'background-color': backgroundColor, color: color }"
      :disabled="disabled"
      @click="click"
    >
      <div class="button__content">
        <div
          class="button__icon button__icon__left"
          :style="{ fill: color }"
          v-if="$slots.left"
        >
          <slot name="left" />
        </div>
        <div class="button__text">
          <slot />
        </div>
        <div
          class="button__icon button__icon__right"
          :style="{ fill: color }"
          v-if="$slots.right"
        >
          <slot name="right" />
        </div>
      </div>
    </component>
  </div>
</template>

<style scoped>
.button {
  border: none;
  font-family: var(--ds-font-family, inherit);
  font-size: initial;
  margin: 0;
  border-radius: 0.35em;
  cursor: pointer;
  user-select: none;
  outline: none;
  transition: all 0.25s;
  text-align: center;
  text-transform: uppercase;
  font-weight: 600;
  letter-spacing: 0.02em;
  text-decoration: none;
  padding: 0;
  width: 100%;
  display: flex;
  justify-content: center;
}
.button:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
.button__content {
  white-space: nowrap;
  display: grid;
  grid-auto-flow: column;
  align-items: center;
}
.button__text {
  padding-left: 0.25rem;
  padding-right: 0.25rem;
}
.button__icon {
  display: flex;
  align-items: center;
}
.button__type__text .button__text {
  padding-left: 0;
  padding-right: 0;
}
.button__size__s {
  font-size: 0.8125rem;
  line-height: 1rem;
  padding-top: 0.5rem;
  padding-bottom: 0.5rem;
}
.button__size__s.button__type__contained {
  padding-left: 0.75rem;
  padding-right: 0.75rem;
}
.button__size__s .button__content {
  gap: 0.25rem;
}
.button__size__s .button__icon {
  width: 1rem;
  height: 1rem;
}
.button__size__m {
  font-size: 1rem;
  line-height: 1.5rem;
  padding-top: 0.75rem;
  padding-bottom: 0.75rem;
}
.button__size__m .button__content {
  gap: 0.5rem;
}
.button__size__m .button__icon {
  width: 1.5rem;
  height: 1.5rem;
}
.button__size__m.button__type__contained {
  padding-left: 1rem;
  padding-right: 1rem;
}
.button__size__l {
  font-size: 1.25rem;
  line-height: 2rem;
  padding-top: 1rem;
  padding-bottom: 1rem;
}
.button__size__l .button__content {
  gap: 0.75rem;
}
.button__size__l .button__icon {
  width: 2rem;
  height: 2rem;
}
.button__size__l.button__type__contained {
  padding-left: 1.5rem;
  padding-right: 1.5rem;
}
</style>

<script>
export default {
  props: {
    size: {
      type: String,
      default: "m",
    },
    /**
     * disabled
     */
    disabled: {
      type: Boolean,
      default: false,
    },
    backgroundColor: {
      type: String,
      default: "#5064FB",
    },
    color: {
      type: String,
      default: "white",
    },
    /**
     * type
     */
    type: {
      type: String,
      default: "contained",
    },
    /**
     * tag
     */
    tag: {
      type: String,
      default: "button",
    },
    /**
     * href
     */
    href: {
      type: String,
      default: null,
    },
    /**
     * target
     */
    target: {
      type: String,
      default: null,
    },
  },
  methods: {
    click() {
      this.$emit("click", true);
    },
  },
};
</script>
