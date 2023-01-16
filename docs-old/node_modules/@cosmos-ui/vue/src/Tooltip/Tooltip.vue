<template>
  <div style="display: inline-block; position: relative;" ref="container">
    <button class="term" tabindex="0" @click="select" ref="term">
      <slot></slot>
    </button>
    <div class="tooltip" v-if="show" ref="tooltip" :style="{'--width': width, '--left': left, '--right': right}" tabindex="1" @focus="setPosition($event)">
      <div class="tooltip__wrapper" v-html="definition"></div>
    </div>
  </div>
</template>

<style scoped>
button {
  background: none;
  border: none;
  font-size: inherit;
  font-family: inherit;
  padding: 0;
}

.term {
  cursor: pointer;
}

.term:active {
  outline: none;
}

.tooltip {
  position: absolute;
  width: var(--width);
  left: initial;
  right: var(--right);
  font-size: 0.75rem;
  line-height: 1.5;
  padding-right: 1rem;
  opacity: 0;
  pointer-events: none;
  outline: none;
  transition: all 0.1s ease-in;
  transform: translateY(-1em);
  z-index: 1000000;
  box-sizing: border-box;
}

.tooltip__wrapper {
  padding: 0.75em 1em;
  background: var(--white-100, white);
  box-shadow: 0 0.25em 1.5em rgba(0, 0, 0, 0.15);
  border-radius: 0.5em;
}

.tooltip:focus {
  transform: translateY(0);
  opacity: 1;
  pointer-events: all;
}

@media screen and (max-width: 600px) {
  .tooltip {
    border-radius: 0;
    width: 100vw;
    left: var(--left);
    padding: 0 1em;
  }
}
</style>

<script>
import dict from "./dict.json";
import MarkdownIt from "markdown-it";

export default {
  props: {
    value: {
      type: String
    }
  },
  data: function() {
    return {
      width: "300px",
      left: "0px",
      right: "0px",
      screenWidth: null,
      show: null
    };
  },
  mounted() {
    this.setPosition();
    window.addEventListener("resize", this.setPosition);
  },
  computed: {
    definition() {
      return new MarkdownIt().render(dict[this.value]);
    }
  },
  methods: {
    select() {
      this.show = true;
      this.$nextTick(() => this.$refs.tooltip.focus());
    },
    setPosition(e) {
      this.screenWidth = window.outerWidth;
      if (this.$refs.container) {
        const el = this.$refs.container.getBoundingClientRect();
        if (this.screenWidth - el.left < parseInt(this.width)) {
          this.right = -(this.screenWidth - el.width - el.left) + "px";
          this.left = "initial";
        } else {
          this.right = "initial";
          this.left = -el.left + "px";
        }
      }
    }
  }
};
</script>