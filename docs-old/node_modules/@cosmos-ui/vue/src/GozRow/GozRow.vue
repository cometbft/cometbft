<template>
  <div>
    <div :class="['container', `progress__${!!progress}`]">
      <component :is="url ? 'a' : 'div'" class="row" :href="url" target="_blank" rel="noreferrer noopener">
        <div class="icon__wrapper" v-if="$slots.icon">
          <div class="icon">
            <slot name="icon"/>
          </div>
        </div>
        <div class="details">
          <div class="title">
            <div class="h1">
              <slot name="h1">
                Title
              </slot>
            </div>
            <div class="h2" v-if="$slots.h2">
              <slot name="h2"/>
            </div>
            <div class="body" v-if="$slots.body">
              <slot name="body"/>
            </div>
          </div>
          <div class="indicator">
            <div class="progress__wrapper" v-if="progress">
              <div class="progress" :style="{'--progress-bar-width': `${progress}%`}">
                <div class="progress__bar"></div>
              </div>
              <div class="h3">{{progress}}% <span class="h3__label">complete</span></div>
            </div>
            <div class="aside__icon__wrapper">
              <icon-chevron class="aside__icon"/>
            </div>
          </div>
        </div>
      </component>
    </div>
  </div>
</template>

<style scoped>
.container {
  overflow-wrap: anywhere;
  margin-left: 1rem;
  margin-right: 1rem;
}
.row {
  background-color: var(--grey-23, rgb(46, 49, 72));
  color: var(--white-100, white);
  font-family: var(--ds-font-family, inherit);
  display: grid;
  grid-auto-flow: column;
  grid-template-columns: min-content 1fr;
  text-decoration: none;
  border-radius: .5rem;
  transition: all .25s;
}
.row:hover, .row:focus {
  background-color: #373a52;
  box-shadow: var(--ds-elevation-4);
}
.row:hover .icon, .row:focus .icon {
  opacity: .8;
}
.row:hover .h2, .row:focus .h2 {
  color: white;
}
.row:hover .aside__icon, .row:focus .aside__icon {
  stroke: rgba(255,255,255,.8);
}
.row:active {
  opacity: .8;
}
.row:active .aside__icon {
  opacity: inherit;
}
.icon__wrapper {
  display: flex;
  align-items: center;
}
.icon {
  grid-column-start: 1;
  width: 4rem;
  height: 4rem;
  fill: var(--white-100, white);
  opacity: .32;
  margin: 1.5rem 3rem;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: opacity .25s;
}
.details {
  padding-top: 1.5rem;
  padding-bottom: 1.5rem;
  grid-column-start: 2;
  display: grid;
  grid-auto-flow: column;
  align-items: center;
  gap: 1rem;
  justify-content: space-between;
  margin-left: 1.5rem;
  margin-right: 1.5rem;
}
.h1 {
  font-weight: 500;
  font-size: 1.25rem;
  line-height: var(--ds-h5-line-height);
  letter-spacing: -0.01em;
  color: var(--white-100, white);
  display: block;
  text-decoration: none;
}
.h2 {
  font-size: var(--ds-body2-font-size, .875rem);
  line-height: var(--ds-body2-line-height, 1.25rem);
  line-height: 24px;
  color: var(--white-51);
  font-weight: 400;
  transition: color .25s;
}
.body {
  margin-top: .75rem;
  line-height: var(--ds-body1-line-height, 1.5rem);
  color: rgba(255,255,255,.8);
  font-size: var(--ds-body1-font-size, 1rem);
}
.h3 {
  font-size: .875rem;
  line-height: 20px;
  letter-spacing: 0.01em;
  color: var(--white-80);
  text-align: right;
  text-transform: none;
  font-weight: 400;
}
.indicator {
  display: flex;
  flex-direction: row;
  align-items: flex-end;
}
.progress__wrapper {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
}
.progress {
  position: relative;
  height: .25rem;
  width: 7rem;
  background: var(--white-20);
  margin-bottom: .75rem;
  border-radius: .5rem;
}
.progress__bar {
  height: 100%;
  width: var(--progress-bar-width, 0);
  background: var(--success);
  border-radius: inherit;
}
.aside__icon__wrapper {
  padding: 1rem .75rem;
  margin-left: .5rem;
}
.aside__icon {
  height: 1rem;
  width: auto;
  display: block;
  stroke: rgba(255,255,255,.32);
  transition: all .25s;
}
@media screen and (max-width: 600px) {
  .icon {
    display: none;
  }
  .h1 {
    font-size: var(--ds-body2-font-size, .875rem);
  }
  .h2 {
    font-size: var(--ds-caption-book-font-size, 0.8125rem);
  }
  .details {
    margin-left: 1rem;
    margin-right: 1rem;
    padding-top: 1rem;
    padding-bottom: 1rem;
  }
  .progress__true .aside__icon__wrapper {
    display: none;
  }
  .progress {
    width: 3rem;
  }
  .h3__label {
    display: none;
  }
}
</style>

<script>
import IconChevron from "./../Icons/IconChevron"

export default {
  components: {
    IconChevron
  },
  props: {
    progress: {
      type: [String, Number]
    },
    url: {
      type: String
    }
  },
}
</script>