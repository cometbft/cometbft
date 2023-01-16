<template>
  <div>
    <div class="wrapper">
      <div class="container" :style="{'background': `linear-gradient(180deg, rgba(25, 52, 167, 0.9) 0%, rgba(14, 14, 86, 0.9) 100%), url(${imgSrc})`, 'background-size': 'cover', 'background-repeat': 'no-repeat', 'background-position': 'center center'}">
        <div class="h1">Compete to win 100,000 ATOM</div>
        <div class="p">May the odds be ever in your favor…</div>
        <div class="footer">
          <text-button tag="a" backgroundColor="#9214B1" href="https://forms.gle/umjGa9G3Q6hcq2iF7" target="_blank" size="m">
            Sign up
          </text-button>
        </div>
      </div>
      <div class="container" :style="{'background': `linear-gradient(180deg, rgba(46, 49, 72, 0.9) 0%, rgba(22, 25, 49, 0.81) 100%), url(${imgSrc})`, 'background-size': 'cover', 'background-repeat': 'no-repeat', 'background-position': 'center center'}">
        <div class="h1">Join the discussion</div>
        <div class="p">Get involved in the Game of Zones chat room.</div>
        <div class="footer">
          <text-button tag="a" href="https://discord.gg/cr7N47p" target="_blank" size="m">
            Discuss GoZ
          </text-button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.wrapper {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1.5rem;
}
.container {
  font-family: var(--ds-font-family, inherit);
  padding: 4rem 1rem;
  box-shadow: 0px 10px 20px rgba(0, 0, 0, 0.05), 0px 2px 6px rgba(0, 0, 0, 0.05), 0px 1px 0px rgba(0, 0, 0, 0.05);
  border-radius: 0.75rem;
}
.h1 {
  font-weight: 500;
  font-size: 2rem;
  line-height: 2.5rem;
  text-align: center;
  letter-spacing: -0.04em;
  color: var(--white-100, white);
  margin-bottom: 2rem;
}
.p {
  font-size: 1.25rem;
  line-height: 1.75rem;
  text-align: center;
  color: var(--white-80);
}
.footer {
  margin-top: 3rem;
  display: flex;
  justify-content: center;
}
@media screen and (max-width: 600px) {
  .wrapper {
    grid-template-columns: 1fr;
  }
  .container {
    padding: 3rem 1rem;
  }
  .h1 {
    font-size: 1.5rem;
  }
  .p {
    font-size: 1.05rem;
  }
  .footer {
    margin-top: 2rem;
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
      default: "/goz.jpg"
    }
  }
}
</script>