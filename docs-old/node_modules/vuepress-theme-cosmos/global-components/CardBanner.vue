<template lang="pug">
  div
    .container
      icon-arrow.icon
      .image
        svg(width="41" height="44" viewBox="0 0 41 44" fill="none" xmlns="http://www.w3.org/2000/svg")
          path(fill-rule="evenodd" clip-rule="evenodd" d="M4 1.8999H32C33.3807 1.8999 34.5 3.01919 34.5 4.3999V10.997L36 9.497V4.3999C36 2.19076 34.2091 0.399902 32 0.399902H4C1.79086 0.399902 0 2.19077 0 4.39991V39.5999C0 41.809 1.79086 43.5999 4 43.5999H32C34.2091 43.5999 36 41.809 36 39.5999V20.2096L34.5 21.7096V39.5999C34.5 40.9806 33.3807 42.0999 32 42.0999H4C2.61929 42.0999 1.5 40.9806 1.5 39.5999V4.39991C1.5 3.01919 2.61929 1.8999 4 1.8999ZM7.2 19.1999H26.2971L24.2971 21.1999H7.2V19.1999ZM7.2 15.7999H28.8V13.7999H7.2V15.7999ZM7.2 26.5999H16.2V24.5999H7.2V26.5999ZM20.8638 25.7372L20.7386 25.8624L20.6826 26.0303L18.9855 31.1215L18.4792 32.6406L19.9704 32.0571L24.8494 30.1479L24.9955 30.0908L25.1064 29.9798L37.9799 17.1063L39.7831 15.3032C40.9546 14.1316 40.9546 12.2321 39.7831 11.0606C38.6115 9.88898 36.712 9.88898 35.5404 11.0606L20.8638 25.7372ZM20.9149 30.0767L22.0496 26.6726L36.6011 12.1212C37.1869 11.5354 38.1366 11.5354 38.7224 12.1212C39.3082 12.707 39.3082 13.6567 38.7224 14.2425L37.4496 15.5153L36.2465 14.3122C35.9536 14.0193 35.4787 14.0193 35.1858 14.3122C34.8929 14.6051 34.8929 15.08 35.1858 15.3729L36.3889 16.576L24.5761 28.3888L23.6215 27.4342C23.3286 27.1413 22.8537 27.1413 22.5608 27.4342C22.2679 27.7271 22.2679 28.202 22.5608 28.4949L23.2349 29.1689L20.9149 30.0767Z" fill="var(--color-primary, blue)")
      .h1 Found an Issue?
      .h2 Help us improve this page by suggesting edits on GitHub.
</template>

<style lang="stylus" scoped>
.container
  background #f8f9fc
  padding 1rem
  text-align center
  border-radius 0.5rem
  position relative
  transition all 0.1s

  &:hover:not(:active)
    box-shadow 0px 12px 24px rgba(0, 0, 0, 0.07), 0px 4px 8px rgba(0, 0, 0, 0.05), 0px 1px 0px rgba(0, 0, 0, 0.05)
    transform translateY(-2px)
    transition transform 0.15s ease-out, box-shadow 0.15s ease-out

    .icon
      fill rgba(0,0,0,0.4)

  &:active
    transition-duration 0s

.icon
  position absolute
  padding 0.75rem
  width 1rem
  height 1rem
  top 0
  right 0
  fill rgba(0,0,0,0.2)

.image
  display flex
  justify-content center
  padding 0.5rem

.h1
  margin-top 0.875rem
  margin-bottom 0.25rem
  line-height 1.25rem
  font-weight 600

.h2
  font-size 0.8125rem
  line-height 1.125rem
  color rgba(22, 25, 49, 0.65)
</style>
