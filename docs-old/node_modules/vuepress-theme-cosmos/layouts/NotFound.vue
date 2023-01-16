<template lang="pug">
  div
    img( src="./images/lost_cosmonaut.jpg")
    center
      .h1 Error 404 : You seem lost fellow cosmonaut,
      .h1 Let us take you #[a(href="/") Back to Home]
</template>

<style lang="stylus" scoped>
img
  width 100%
  margin-bottom 1rem
  margin-top: 1rem
a
  color var(--color-link, blue)
  outline-color var(--color-primary, blue)
  border-radius 0.25rem
  transition opacity 0.3s ease-out

  &:hover
    text-decoration underline

  &:active
    opacity 0.65
    transition-duration 0s

.h1
  font-size 1.5rem
  line-height 2.5rem
  letter-spacing -0.02em
  font-weight 700

.p {
  font-size 1.125rem
  line-height 1.8125rem
  margin-top 1em
  margin-bottom 1em
}
</style>
