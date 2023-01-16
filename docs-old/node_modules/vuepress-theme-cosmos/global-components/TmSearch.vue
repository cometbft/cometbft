<template lang="pug">
  div
    .container
      .search
        svg(width="25" height="25" viewBox="0 0 25 25" fill="none" xmlns="http://www.w3.org/2000/svg").search__icon
          path(fill-rule="evenodd" clip-rule="evenodd" d="M14.5404 18.8025C18.6351 17.0622 20.5438 12.332 18.8036 8.23728C17.0634 4.14253 12.3332 2.2338 8.23843 3.97403C4.14367 5.71426 2.23495 10.4445 3.97518 14.5392C5.71541 18.634 10.4456 20.5427 14.5404 18.8025Z" stroke="#A2A3AD" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round")
          path(d="M17.0854 17.0845L23.3334 23.3334" stroke="#A2A3AD" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round")
        input(placeholder="Search" @input="$emit('input', $event.target.value)").search__input
</template>

<style lang="stylus" scoped>
.container
  margin-top 2rem
  margin-bottom 2rem

.search
  display flex
  position relative

  &__icon
    position absolute
    top .25rem
    left .5rem
    pointer-events none
  
  &__input
    border 1px solid rgba(176, 180, 207, 0.09)
    font-family inherit
    font-size .875rem
    padding .5rem 1.5rem
    padding-left 2.5rem
    width 100%
    background rgba(176, 180, 207, 0.09)
    border-radius 10000px
    outline none
    transition all .25s

    &:focus
      border 1px solid var(--color-primary, blue)
      background white
</style>

<script>
export default {
  props: ["value"]
};
</script>