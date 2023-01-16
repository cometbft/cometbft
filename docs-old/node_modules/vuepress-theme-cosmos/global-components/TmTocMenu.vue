<template lang="pug">
  div
    .menu__item__modal__title On this page
    .menu__item__modal__item(v-for="item in $page.headers" v-if="!item.title.match(/{hide}/i)" @click="goto(item.slug)" tabindex="0") {{item.title}}
</template>

<style lang="stylus" scoped>
.menu
  visibility hidden

  &__item
    position relative

    &__icon
      cursor pointer
      width 1.5rem
      height 1.5rem
      padding .75rem
      border-radius .25rem
      fill rgba(51, 54, 74, 0.4)

      &.menu__item__icon__active__true
        fill #5064FB
        background rgba(102, 161, 255, .15)

    &__modal
      position absolute
      width 16rem
      box-shadow 0px 24px 40px rgba(22, 25, 49, 0.1), 0px 10px 16px rgba(22, 25, 49, 0.08), 0px 1px 0px rgba(22, 25, 49, 0.05)
      right 0
      z-index 1000
      border-radius .25rem
      background-color white

      &__title
        color rgba(22, 25, 49, 0.65)
        text-transform uppercase
        letter-spacing 0.2em
        font-size .75rem
        padding 1.5rem 1.5rem 1rem 1.5rem

      &__item
        padding .625rem 1.5rem
        font-size .875rem
        line-height 1.25rem
        outline none
        cursor pointer

        &:hover
          background-color rgba(176, 180, 207, 0.087)
          color var(--color-text, black)

        &:active
          color var(--color-link)
          font-weight 600
          box-shadow inset 2px 0 0 0 var(--color-link)

        &:last-child
          margin-bottom 0.5rem
</style>

<script>
export default {
  methods: {
    goto(fragment) {
      location.replace(`#${fragment}`);
      location.replace(`#${fragment}`);
      this.$emit("visible", false);
      this.$emit("toc", false);
    }
  }
};
</script>
