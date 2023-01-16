<template>
  <div>
    <div
      :class="['container', `fullscreen__${!!fullscreen}`]"
      :style="{
        '--page-min-height': pageMinHeight,
        background: background || 'none',
      }"
    >
      <background-stars v-if="!background" />
      <div class="wrapper">
        <div class="image">
          <div class="image__img" v-if="step === 2" key="i1">
            <graphics-mail class="image__img__img" />
          </div>
          <div class="image__img" v-else key="i2">
            <graphics-planes class="image__img__img" />
          </div>
        </div>
        <div class="text">
          <transition-group
            class="page__container"
            :name="transition"
            @before-enter="setHeight"
          >
            <div class="page" v-show="step === 0" ref="step0" key="step0">
              <div class="page__wrapper">
                <div class="icon-hero" v-if="iconHero">
                  <div v-html="iconHero"></div>
                </div>
                <div class="h4" v-if="fullscreen">Email communications</div>
                <label for="newsletter_email" class="h1">{{ h1 }}</label>
                <div class="p1">{{ h2 }}</div>
                <div class="email__form">
                  <div class="email__form__input">
                    <input
                      @keypress.enter="actionSubmitEmail"
                      id="newsletter_email"
                      v-model="email"
                      class="email__form__input__input"
                      type="text"
                      placeholder="Your email"
                    />
                  </div>
                  <ds-button
                    class="button-sign-up"
                    @click="actionSubmitEmail"
                    :disabled="emailInvalid"
                  >
                    Sign up
                    <template v-slot:right>
                      <icon-arrow-right />
                    </template>
                  </ds-button>
                </div>
                <div class="p2">
                  You can unsubscribe at any time.
                  <a href="https://cosmos.network/privacy" target="_blank"
                    >Privacy Policy</a
                  >
                </div>
              </div>
            </div>
            <div class="page" v-show="step === 1" ref="step1" key="step1">
              <div class="page__wrapper">
                <div style="display: inline-block">
                  <ds-button
                    size="s"
                    color="#66A1FF"
                    backgroundColor="rgba(0,0,0,0)"
                    type="text"
                    @click.native="actionGoBackwards"
                  >
                    <template v-slot:left>
                      <icon-chevron-left />
                    </template>
                    Back
                  </ds-button>
                </div>
                <div class="h2">What are you interested in?</div>
                <div class="card-checkbox-list">
                  <card-checkbox
                    v-for="(topic, i) in topics"
                    :key="topic.h1"
                    v-model="selected[i]"
                    theme="dark"
                  >
                    <template v-slot:icon>
                      <div v-if="icons[i]" v-html="icons[i]"></div>
                      <icon-network v-else />
                    </template>
                    <template v-slot:h1>
                      {{ topic.h1 }}
                    </template>
                    {{ topic.h2 }}
                  </card-checkbox>
                </div>
                <div style="display: inline-block">
                  <ds-button
                    size="l"
                    @click="actionSubscribe"
                    :disabled="!selected.some((t) => t)"
                  >
                    Get updates
                  </ds-button>
                </div>
              </div>
            </div>
            <div class="page" v-show="step === 2" ref="step2" key="step2">
              <div class="page__wrapper">
                <div class="h1">Almost there…</div>
                <div class="p1">
                  You should get a confirmation email for each of your selected
                  interests. Open it up and click ‘<strong
                    >Confirm your email</strong
                  >’ so we can keep you updated.
                </div>
                <div class="h3">Don’t see the confirmation email yet?</div>
                <div class="p2">
                  It might be in your spam folder. If so, make sure to mark it
                  as “not spam”.
                </div>
              </div>
            </div>
          </transition-group>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
a {
  color: #66a1ff;
  text-decoration: none;
}
.icon-hero {
  width: 4rem;
  height: 4rem;
  margin: 1.5rem 0 2rem;
  color: rgba(255, 255, 255, 0.5);
}
.container {
  background: var(--newsletter-background);
  font-family: var(--ds-font-family, inherit);
  color: white;
  position: relative;
}
.wrapper {
  display: grid;
  grid-template-columns: 50% 50%;
  grid-template-rows: 1fr;
  align-items: center;
  width: 100%;
}
.image {
  width: 100%;
  height: 100%;
  position: relative;
  overflow: hidden;
  transform: translateZ(0);
}
.image__img {
  position: absolute;
  top: 50%;
  left: 100%;
  transform: translate(-85%, -50%);
}
.text {
  max-width: 36rem;
  position: relative;
  width: 100%;
  overflow-y: hidden;
  transition: min-height 0.5s ease-in-out;
  height: 100%;
  min-height: var(--page-min-height);
}
.page {
  padding-left: 1rem;
  padding-right: 1rem;
  box-sizing: border-box;
  padding-top: 5rem;
  padding-bottom: 5rem;
  width: 100%;
  display: flex;
  align-items: center;
}
.page__wrapper {
  width: 100%;
}
.page__container {
  display: flex;
  align-items: center;
  height: 100%;
}
.h1 {
  font-size: 2rem;
  font-weight: 600;
  line-height: 1.25;
  color: var(--white-100, white);
  margin-bottom: 1.5rem;
  display: block;
}
.fullscreen__true .h1 {
  font-size: 2.5rem;
}
.h2 {
  font-size: 2rem;
  color: white;
  font-weight: 600;
  line-height: 1.25;
  margin-top: 2.5rem;
  margin-bottom: 2.5rem;
}
.h3 {
  font-size: 0.8125rem;
  font-weight: 600;
  color: rgba(255, 255, 255, 0.8);
  text-transform: none;
  margin: initial;
  letter-spacing: initial;
  line-height: 1.25rem;
}
.h4 {
  margin-bottom: 0.5rem;
  font-weight: 800;
  line-height: 20px;
  letter-spacing: 0.07em;
  text-transform: uppercase;
  color: rgba(255, 255, 255, 0.5);
}
.p1 {
  color: rgba(255, 255, 255, 0.8);
  font-size: 1.25rem;
  line-height: 1.4;
  font-weight: 400;
  margin-bottom: 2rem;
}
.p2 {
  font-size: 0.8125rem;
  color: rgba(255, 255, 255, 0.5);
  line-height: 1.25rem;
}
.email__form {
  display: grid;
  grid-auto-flow: column;
  grid-template-columns: 1fr min-content;
  gap: 1rem;
  margin-top: 2.5rem;
  margin-bottom: 1.5rem;
}
.email__form__input__input {
  outline: none;
  width: 100%;
  background: rgba(255, 255, 255, 0.1);
  border: none;
  border-radius: 0.25rem;
  color: white;
  font-size: 1rem;
  padding: 0.75rem 1rem;
  line-height: 1.5;
  height: auto;
  box-sizing: border-box;
  transition: all 0.15s;
  backdrop-filter: blur(10px);
  font-family: var(--ds-font-family, inherit);
  font-weight: 400;
  margin: 0;
}
.email__form__input__input:hover {
  background-color: rgba(255, 255, 255, 0.13);
}
.email__form__input__input:focus {
  box-shadow: inset 0 0 0 1.5px #66a1ff;
  background-color: rgba(0, 0, 0, 0.2);
}
.email__form__input__input::-webkit-input-placeholder,
.email__form__input__input::placeholder {
  color: rgba(255, 255, 255, 0.5);
  transition: color 0.15s;
}
.email__form__input__input:hover:not(:focus)::-webkit-input-placeholder,
.email__form__input__input:hover:not(:focus)::placeholder {
  color: rgba(255, 255, 255, 0.8);
}
.card-checkbox-list {
  margin-top: 2.5rem;
  margin-bottom: 2.5rem;
  display: grid;
  grid-auto-flow: row;
  gap: 1rem;
}
.forwards-enter-active,
.forwards-leave-active,
.backwards-enter-active,
.backwards-leave-active {
  transition: all 0.5s ease-in-out;
  position: absolute;
}
.forwards-enter {
  opacity: 0;
  transform: translateY(50px);
}
.forwards-enter-to {
  opacity: 1;
  transform: translateY(0);
}
.forwards-leave {
  opacity: 1;
  transform: translateY(0);
}
.forwards-leave-to {
  opacity: 0;
  transform: translateY(-50px);
}
.backwards-enter {
  opacity: 0;
  transform: translateY(-50px);
}
.backwards-enter-to {
  opacity: 1;
  transform: translateY(0);
}
.backwards-leave {
  opacity: 1;
  transform: translateY(0);
}
.backwards-leave-to {
  opacity: 0;
  transform: translateY(50px);
}
.fade-enter-active,
.fade-leave-active {
  transition: opacity 5s;
}
.fade-enter,
.fade-leave-to {
  opacity: 0;
}
.fade-enter-to,
.fade-leave {
  opacity: 1;
}
@media screen and (max-width: 800px) {
  .icon-hero {
    display: flex;
    justify-content: center;
    width: 100%;
  }
  .wrapper {
    grid-template-columns: 1fr;
  }
  .page__container {
    align-items: flex-start;
  }
  .page {
    padding-bottom: 0;
  }
  .image {
    height: 250px;
    display: flex;
    justify-content: center;
  }
  .image__img {
    position: relative;
    left: 0;
    top: 0;
    transform: translateY(-100%);
  }
  .image__img__img {
    width: 500px;
    left: 50%;
    position: absolute;
    transform: translateX(-50%);
  }
  .text {
    transform: translateY(-75px);
    justify-self: center;
    max-width: 560px;
    height: auto;
  }
  .h1,
  .p1,
  .p2,
  .h3,
  .h4 {
    text-align: center;
  }
}
@media screen and (max-width: 400px) {
  .email__form {
    grid-template-columns: 1fr;
    grid-auto-flow: row;
  }
  .h1,
  .h2,
  .fullscreen__true .h1 {
    font-size: 1.5rem;
  }
  .h2 {
    margin-top: 1rem;
    margin-bottom: 1rem;
  }
  .p1 {
    font-size: 1rem;
  }
  .h4 {
    font-size: 0.8125rem;
  }
}
</style>

<script>
import GraphicsPlanes from "./GraphicsPlanes";
import GraphicsMail from "./GraphicsMail";
import DsButton from "./DsButton";
import IconArrowRight from "../Icons/IconArrowRight";
import IconChevronLeft from "../Icons/IconChevronLeft";
import IconNetwork from "../Icons/IconNetwork";
import IconIbc from "../Icons/IconIbc";
import CardCheckbox from "./CardCheckbox";
import querystring from "querystring";
import axios from "axios";
import BackgroundStars from "./BackgroundStars";

export default {
  props: {
    h1: {
      default: "Sign up for Cosmos updates"
    },
    h2: {
      default:
        "Get the latest from the Cosmos ecosystem and engineering updates, straight to your inbox."
    },
    requestURL: {
      default: "https://app.mailerlite.com/webforms/submit/o0t6d7"
    },
    callback: {
      default: "jQuery18307296239382192573_1594158619276"
    },
    _: {
      default: "1594158625563"
    },
    groups: {
      default: false
    },
    svg: {
      default: false
    },
    background: {
      default: false
    },
    topics: {
      default: () => []
    },
    banner: {
      type: Boolean,
      default: true
    },
    fullscreen: {
      type: String
    }
  },
  watch: {
    fullscreen() {
      this.setHeight(`step${this.step}`);
    }
  },
  components: {
    GraphicsPlanes,
    GraphicsMail,
    DsButton,
    IconArrowRight,
    IconChevronLeft,
    IconIbc,
    CardCheckbox,
    IconNetwork,
    BackgroundStars
  },
  data: function() {
    return {
      step: 0,
      transition: "forwards",
      pageMinHeight: null,
      email: null,
      selected: this.topics.map(t => false),
      icons: [],
      ready: false,
      iconHero: false,
      commonFormData: {
        "ml-submit": "1",
        "ajax": "1",
        "guid": "6ca22b31-4124-e926-cf4f-272ff9f44ec3"
      }
    };
  },
  async mounted() {
    if (this.svg) {
      this.iconHero = (await axios.get(this.svg)).data;
    }
    this.setHeight(`step${this.step}`);
    this.ready = true;
    this.topics.forEach(async topic => {
      let icon = false;
      try {
        icon = (await axios.get(topic.svg)).data;
      } catch {
        console.error(`Can't load icon from ${topic.svg}.`);
      }
      this.icons.push(icon);
    });
  },
  computed: {
    emailInvalid() {
      const re = /^(([^<>()\[\]\\.,;:\s@"]+(\.[^<>()\[\]\\.,;:\s@"]+)*)|(".+"))@((\[[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\])|(([a-zA-Z\-0-9]+\.)+[a-zA-Z]{2,}))$/;
      return !re.test(String(this.email));
    }
  },
  methods: {
    actionSubmitEmail() {
      if (!this.emailInvalid) {
        if (this.topics.length <= 0) {
          this.actionSubscribe();
          this.step = 2;
        } else {
          this.actionGoForwards();
        }
      }
    },
    actionSubscribe(selected) {
      if (this.topics.length <= 0) {
        this.subscribe({
          requestURL: this.requestURL,
          callback: this.callback,
          _: this._,
          "groups[]": this.groups
        });
      } else {
        this.selected.forEach((topicSelected, i) => {
          if (topicSelected) {
            this.subscribe({
              requestURL: this.topics[i].requestURL,
              callback: this.topics[i].callback,
              _: this.topics[i]._,
              "groups[]": this.topics[i].groups
            });
          }
        });
      }
      this.actionGoForwards();
    },
    setHeight(el) {
      this.$nextTick(() => {
        const isString = s => typeof s === "string" || s instanceof String;
        const page = isString(el) ? this.$refs[el] : el;
        const height =
          this.fullscreen && window.innerWidth > 800
            ? this.fullscreen
            : page.getBoundingClientRect().height + "px";
        this.pageMinHeight = height;
      });
    },
    actionGoForwards() {
      this.transition = "forwards";
      this.step += 1;
    },
    actionGoBackwards() {
      this.transition = "backwards";
      this.step -= 1;
    },
    async subscribe(body) {
      const options = {
        method: "POST",
        mode: "no-cors",
        headers: {
          "Content-Type": "application/x-www-form-urlencoded"
        },
        body: querystring.stringify({
          "fields[email]": this.email,
          ...this.commonFormData,
          ...body
        })
      };
      fetch(this.requestURL, options);
    }
  }
};
</script>
