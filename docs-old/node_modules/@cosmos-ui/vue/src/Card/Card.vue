<template>
  <div>
    <div class="card" :style="cardStyle">
      <div :class="['image', `image__side__${imgSide}`]">
        <img class="image__img" v-if="imgSrc" :src="imgSrc" alt="">
      </div>
      <div class="text">
        <div v-if="overline" class="overline">
          {{overline}}
        </div>
        <div v-if="title" class="title">
          {{title}}
        </div>
        <div :class="['body', `body__canexpand__${!!(canExpand)}`, `body__expanded__${!!(isExpanded)}`]" ref="body" @transitionend="transitionend">
          <!-- @slot Main content of the card -->
          <slot/>
          <transition name="fade">
            <div class="body__more" @click="expand" v-if="canExpand && !isExpanded">read more</div>
          </transition>
        </div>        
      </div>
    </div>
  </div>
</template>

<style scoped>
.card {
  display: flex;
  flex-direction: var(--card-grid-auto-flow);
  font-family: var(--ds-font-family, inherit);
  border-radius: var(--card-border-radius, 8px);
  position: relative;
  background: var(--white-100, white);
}
.card:before {
  border-radius: var(--card-border-radius, 8px);
  content: "";
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  z-index: -1;
  box-shadow: var(--ds-elevation-2);
}
.image {
  flex-shrink: 0;
}
.image__side__top {
  width: 100%;
  height: var(--card-image-height);
  overflow: hidden;
  position: relative;
}
.image__side__top img {
  border-top-left-radius: var(--card-border-radius, 8px);
  border-top-right-radius: var(--card-border-radius, 8px);
}
.image__side__left {
  width: var(--card-image-height);
}
.image__side__left img {
  border-top-left-radius: var(--card-border-radius, 8px);
  border-bottom-left-radius: var(--card-border-radius, 8px);
}
.image__side__right {
  width: var(--card-image-height);
  order: 2;
}
.image__side__right img {
  border-top-right-radius: var(--card-border-radius, 8px);
  border-bottom-right-radius: var(--card-border-radius, 8px);
}
.image__img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}
.text {
  padding: calc(var(--card-size) - 4px) var(--card-size);
  order: 1;
}
.title {
  font-size: var(--card-title-font-size, 1rem);
  font-weight: 500;
  line-height: var(--card-title-line-height, 1.5em);
  color: var(--grey-14);
  margin-bottom: var(--card-title-margin-bottom, .25em);
}
.overline {
  font-size: .75rem;
  text-transform: uppercase;
  color: var(--grey-trans-44);
  letter-spacing: 0.2em;
  margin-bottom: .3em;
}
.body {
  color: var(--grey-trans-44);
  font-size: var(--card-body-font-size, 14px);
  line-height: var(--card-body-line-height, 1.5em);
  letter-spacing: 0.01em;
  overflow: hidden;
  position: relative;
  transition: max-height cubic-bezier(0.785, 0.135, 0.15, 0.86) .5s;
}
.body__canexpand__true.body__expanded__false {
  max-height: calc(var(--card-body-line-height) * 3);
}
.body__expanded__true {
  max-height: calc(var(--card-body-max-height));
}
.body__more {
  position: absolute;
  bottom: 0;
  right: 0;
  background: white;
  color: var(--primary, blue);
  cursor: pointer;
}
.body__more::before {
  content: "";
  background-image: linear-gradient(to right, rgba(255,255,255,0), white 80%);
  width: 50px;
  height: 100%;
  position: absolute;
  left: 0;
  transform: translateX(-100%);
  top: 0;
}
.fade-leave-active {
  transition: all .5s;
}
.fade-leave {
  opacity: 1;
}
.fade-leave-to {
  opacity: 0;
}
</style>

<script>
export default {
  props: {
    /**
     * Text at the top
     */
    overline: {
      type: String,
      default: "Overline"
    },
    /**
     * Main title
     */
    title: {
      type: String,
      default: "Title"
    },
    /**
     * Size of the card
     */
    size: {
      type: String,
      default: "16px"
    },
    /**
     * Image URL
     */
    imgSrc: {
      type: String,
      default: ""
    },
    /**
     * Position of the image: `top` | `left` | `right`
     */
    imgSide: {
      type: String,
      default: "top",
      validator: (value) => {
        return ['top', 'left', 'right'].indexOf(value) !== -1
      }
    },
    imgSize: {
      type: String,
      default: "120px"
    }
  },
  data: function() {
    return {
      maxHeight: null,
      lineHeight: null,
      isExpanded: null
    }
  },
  mounted() {
    this.maxHeight = this.$refs.body.scrollHeight
    this.lineHeight = parseInt(getComputedStyle(this.$refs.body)['line-height'])
  },
  methods: {
    expand() {
      this.maxHeight = this.$refs.body.scrollHeight
      this.isExpanded = !this.isExpanded
    },
    transitionend(e) {
      if (this.isExpanded) {
        this.maxHeight = null
      }
    }
  },
  computed: {
    canExpand() {
      /**
       * Prevent `read more` from showing up on component mount.
       */
      if (this.maxHeight && this.lineHeight) {
        return this.maxHeight >= 4 * this.lineHeight
      }
    },
    cardStyle() {
      const keywords = {
        "small": "16px",
        "medium": "24px",
        "large": "32px"
      }
      const size = parseInt(keywords[this.size] || this.size)
      return {
        "--card-size": size + 'px',
        "--card-border-radius": "8px",
        "--card-grid-auto-flow": this.imgSide === "top" ? "column" : "row",
        "--card-image-height": this.imgSize,
        "--card-title-font-size": size <= 18 ? "16px" : "20px",
        "--card-title-line-height": size <= 18 ? "24px" : "28px",
        "--card-title-margin-bottom": size <= 18 && "4px" || size <= 24 && "8px" || "12px",
        "--card-body-font-size": size <= 18 ? "13px" : "14px",
        "--card-body-line-height": size <= 18 ? "18px" : "20px",
        "--card-body-line-height-actual": this.lineHeight + 'px',
        "--card-body-max-height": this.canExpand && this.maxHeight + 'px'
      }
    }
  }
}
</script>