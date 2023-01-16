<template>
  <div class="component__hero__container" v-bind:style="{'background-image': `url(${imgSrc})`, 'background-size': 'cover', 'background-repeat': 'no-repeat', 'background-position': 'center center'}">
    <div class="container">
      <div class="section">
        <img class="section__img" v-bind:src="logoSrc" alt="">
        <div class="footer">
          <div class="subtitle">Compete to <b>win 100,000 ATOM</b> by building and running Cosmos Zones.</div>
          <text-button tag="a" href="https://forms.gle/umjGa9G3Q6hcq2iF7" target="_blank" size="l">
            <div class="button">
              sign up
              <icon-arrow-right class="button__icon"/>
            </div>
          </text-button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.component__hero__container {
  font-family: var(--ds-font-family, inherit);
  position: relative;
  min-height: 100vh;
}
.component__hero__container:before {
  content: "";
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: linear-gradient(180deg, rgba(22,25,49,0) 68.75%, rgba(22,25,49,0.8) 88.02%);
}
.container {
  padding: 14rem 1rem;
  border-radius: 0.75rem;
  margin-left: auto;
  margin-right: auto;
}
.section {
  display: flex;
  justify-content: center;
  align-items: center;
}
.section__img {
  height: 7.5rem;
  width: 100%;
  max-width: 65rem;
  position: absolute;
}
.footer {
  bottom: 3rem;
  position: absolute;
}
.subtitle {
  color: var(--white-100, white);
  font-size: 1.5rem;
  line-height: 2rem;
  text-align: center;
  letter-spacing: -0.01em;
  max-width: 34.4162rem;
  width: 100%;
}
.button {
  display: grid;
  grid-auto-flow: column;
  align-items: center;
  justify-content: center;
  white-space: nowrap;
  gap: .75rem;
  box-sizing: border-box;
  color: var(--white-100, white);
}
.button__style__standard {
  background: var(--secondary-dark);
  display: flex;
  width: fit-content;
  justify-content: center;
  align-items: center;
  margin: 1.5rem auto;
}
.button__style__standard:hover, .button__style__standard:focus {
  background: var(--secondary);
}
.button__icon {
  width: 1.75rem;
  height: 2rem;
  fill: var(--white-100, white);
}
@media screen and (max-width: 600px) {
  .footer {
    margin: auto 1rem;
  }
}
@media screen and (max-width: 414px) {
  .container {
    padding: 12rem 0;
  }
  .footer {
    bottom: 1rem;
  }
  .section__img {
    max-width: 18rem;
  }
}
@media screen and (max-width: 375px) {
  .container {
    padding: 13rem 0;
  }
}
@media screen and (max-width: 320px) {
  .container {
    padding: 10rem 0;
  }
  .subtitle {
    font-size: 1.3rem;
  }
}
</style>

<script>
import Button from "../Button/Button"
import IconArrowRight from "./../Icons/IconArrowRight"

export default {
  components: {
    IconArrowRight,
    "text-button": Button
  },
  props: {
    imgSrc: {
      type: String,
      default: "https://cosmos-network-npfcgtew9.now.sh/img/section-gameofzones.bc705fc3.jpg"
    },
    logoSrc: {
      type: String,
      default: "/goz-logo.svg"
    }
  },
};
</script>