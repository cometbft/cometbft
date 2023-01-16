<template>
  <div>
    <div class="wrapper">
      <div class="container">
        <div class="section">
          <transition name="fade" mode="out-in">
            <div v-if="state === 'success'" key="success">
              <div class="icon">
                <icon-letter-heart class="icon__icon icon__icon__active"/>
              </div>
              <div class="h1">
                Almost there...
              </div>
              <div class="p">
                You should get a confirmation email shortly. Open it up and ‘<strong>Confirm your email</strong>’ to save your spot in the upcoming workshops.
              </div>
              <div class="box">
                <div class="box__h1">
                  Don’t see the confirmation email yet?
                </div>
                <div class="box__p">
                  It might be in your spam folder. If so, make sure to mark it as “not spam”.
                </div>
              </div>
            </div>
            <div v-else-if="state === 'error'" key="error">
              <div class="icon">
                <icon-error class="icon__icon icon__icon__error" v-if="state === 'error'"/>
              </div>
              <div class="h1">
                Uh oh! Something went wrong.
              </div>
              <div class="p">
                Try refreshing the page and submitting your email address again.
              </div>
            </div>
            <div v-else key="default">
              <div class="icon">
                <icon-code class="icon__icon"/>
              </div>
              <div class="h1" v-if="this.$slots['h1']">
                <slot name="h1"/>
              </div>
              <div class="p">
                We'll send you email notifications before each workshop, a link to add the program to your calendar, and handy tips to prepare for the workshops.
              </div>
              <div class="form__wrapper">
                <form :action="url" method="POST" target="_blank" rel="noreferrer noopener" @submit.prevent="submit">
                  <div class="form">
                    <div class="form__input">
                      <input name="CONTACT_EMAIL" v-model="email" class="form__input__input" type="email" placeholder="Your email">
                    </div>
                    <text-button type="submit" :disabled="emailInvalid || requestInFlight" class="form__button" size="m">
                      <div :class="['form__button__content', `form__button__content__in-flight__${!!requestInFlight}`]">
                        sign up
                        <icon-arrow-right class="form__button__icon"/>
                      </div>
                      <div class="form__button__spinner" v-if="requestInFlight">
                        <icon-spinner/>
                      </div>
                    </text-button>
                  </div>
                  <div class="form__p">
                    Zero spam. Unsubscribe at any time. <a href="https://cosmos.network/privacy" target="blank_" rel="noopener noreferrer">Privacy policy</a>
                  </div>
                </form>
              </div>
            </div>
          </transition>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
a {
  color: #5064FB;
  text-decoration: none;
}
.container {
  color: var(--white-100, white);
  font-family: var(--ds-font-family, inherit);
  padding: 4rem 1rem;
  box-shadow: 0px 10px 20px rgba(0, 0, 0, 0.05), 0px 2px 6px rgba(0, 0, 0, 0.05), 0px 1px 0px rgba(0, 0, 0, 0.05);
  margin-left: auto;
  margin-right: auto;
  overflow-x: hidden;
}
.section {
  max-width: 33rem;
  margin-left: auto;
  margin-right: auto;
}
.icon {
  display: flex;
  justify-content: center;
}
.icon__icon {
  stroke: var(--white-100, white);
  width: 4rem;
  height: 4rem;
}
.icon__icon.icon__icon__active {
  stroke: var(--primary-light);
  opacity: 1;
}
.icon__icon.icon__icon__error {
  stroke: var(--danger);
  opacity: 1;
}
.h1 {
  text-align: center;
  font-size: 40px;
  font-weight: 500;
  line-height: 3rem;
  letter-spacing: -0.03em;
  color: #000000;
  margin-top: 1.5rem;
  margin-bottom: 1.5rem;
}
.p {
  font-size: 1.125rem;
  line-height: 1.6875rem;
  text-align: center;
  letter-spacing: -0.01em;
  color: rgba(0, 0, 0, 0.8);
}
.form {
  margin-top: 3rem;
  display: grid;
  gap: 1rem;
  grid-auto-flow: column;
  grid-template-columns: 1fr min-content;
  max-width: 30rem;
  margin-left: auto;
  margin-right: auto;
}
.form__input__input {
  background: none;
  border: 2px solid rgba(59, 66, 125, 0.12);
  padding: .75rem 1rem;
  font-size: 1rem;
  font-family: var(--ds-font-family, inherit);
  line-height: 1.5;
  border-radius: .25rem;
  width: 100%;
  box-sizing: border-box;
  color: rgba(0, 0, 0, 0.667);
  opacity: 0.7;
}
.form__input__input:hover {
  background-color: rgba(255,255,255,0.13);
}
.form__input__input:focus {
  outline: none;
  border: 2px solid #5064FB;
}
.form__button {
  position: relative;
  display: flex;
  justify-content: center;
}
.form__button__content {
  display: grid;
  white-space: nowrap;
  gap: .75rem;
  grid-auto-flow: column;
  align-items: center;
  justify-content: center;
  box-sizing: border-box;
  color: #FFFFFF;
}
.button__style__standard {
  background: #5064FB;
}
.form__button__content.form__button__content__in-flight__true {
  opacity: 0
}
.form__button__spinner {
  width: 1.5rem;
  height: 1.5rem;
  position: absolute;
  animation: spin 1s infinite linear;
}
@keyframes spin {
  from {
    transform: rotate(0deg)
  }
  to {
    transform: rotate(360deg)
  }
}
.form__button__icon {
  fill: #FFFFFF;
  width: 1.5rem;
  height: 1.5rem;
}
.form__input__input::placeholder {
  color: inherit;
  opacity: 0.7;
}
.form__p {
  text-align: center;
  color: rgba(0, 0, 0, 0.667);
  margin-top: 1.5rem;
  font-size: .8125rem;
}
.box {
  box-sizing: border-box;
  border-radius: 8px;
  padding: 1.5rem;
  margin-top: 2.5rem;
  font-size: .875rem;
  line-height: 1.25rem;
  text-align: center;
  letter-spacing: 0.01em;
  margin-bottom: .5rem;
}
.box__h1 {
  font-weight: 500;
  color: rgba(0, 0, 0, 0.667);
}
.box__p {
  color: rgba(0, 0, 0, 0.667);
}
.fade-enter-active {
  transition: all .4s ease-out;
}
.fade-leave-active {
  transition: all .2s ease-out;
}
.fade-enter {
  opacity: 0;
  transform: scale(1.5);
}
.fade-enter-to {
  opacity: 1;
  transform: scale(1);
}
.fade-leave {
  opacity: 1;
  transform: scale(1);
}
.fade-leave-to {
  opacity: 0;
  transform: scale(.85);
}
@media screen and (max-width: 600px) {
  .h1 {
    font-size: 2rem;
    font-weight: 500;
    line-height: 2.5rem;
  }
  .form {
    margin-top: 2rem;
    grid-auto-flow: row;
    grid-template-columns: 1fr;
  }
}
</style>

<script>
import querystring from "querystring"
import IconLetterHeart from "./../Icons/IconLetterHeart"
import IconArrowRight from "./../Icons/IconArrowRight"
import IconPaperPlane from "./../Icons/IconPaperPlane"
import IconError from "./../Icons/IconError"
import IconSpinner from "./../Icons/IconSpinner"
import IconCode from "./../Icons/IconCode"
import Button from "../Button/Button"

export default {
  components: {
    IconLetterHeart,
    IconArrowRight,
    IconPaperPlane,
    IconError,
    IconSpinner,
    IconCode,
    "text-button": Button
  },
  data: function() {
    return {
      email: null,
      state: "default",
      requestInFlight: null,
      url: "https://app.mailerlite.com/webforms/submit/d7i4g7",
      formData: {
        "callback": "jQuery1830520133881537445_1594145870016",
        "ml-submit": "1",
        "ajax": "1",
        "guid": "6ca22b31-4124-e926-cf4f-272ff9f44ec3",
        "_": "1594145875469"
      }
    }
  },
  computed: {
    emailInvalid() {
      const re = /^(([^<>()\[\]\\.,;:\s@"]+(\.[^<>()\[\]\\.,;:\s@"]+)*)|(".+"))@((\[[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\])|(([a-zA-Z\-0-9]+\.)+[a-zA-Z]{2,}))$/;
      return !re.test(String(this.email))
    }
  },
  methods: {
    async submit() {
      this.requestInFlight = true
      const urlParams = new URLSearchParams(window.location.search)
      const options = {
        method: "POST",
        mode: "no-cors",
        headers: {
          "Content-Type": "application/x-www-form-urlencoded"
        },
        body: querystring.stringify({
          "fields[email]": this.email,
          ...this.formData
        })
      }
      fetch(this.url, options).then(_ => {
        this.state = "success"
        this.requestInFlight = false
      })
    }
  },
}
</script>