<template>
  <div class="component__footer__container">
    <slot/>
    <div class="grid">
      <div class="header">
        <div class="header__title">
          <!-- @slot Can be used to add `<img>` with logo to the top left corner of the component. If slot is empty, defaults to `h1` prop value. -->
          <slot name="logo">{{h1}}</slot>
        </div>
        <div class="header__links" ref="links">
          <transition name="fade">
            <div class="header__links__popover" v-if="popover" :style="{'--pos-x': popoverX + 'px', '--pos-y': popoverY + 'px'}">
              <div class="header__links__popover__content">{{popover.title || popover.url || popover}}</div>
            </div>
          </transition>
          <a @mouseover="linkMouseover(link, $event, true)" @mouseleave="linkMouseover(link, $event, false)" :href="url(link)" v-for="link in links" :key="url(link)" class="header__links__item" target="_blank" >
            <svg width="24" height="24" xmlns="http://www.w3.org/2000/svg" fill-rule="evenodd" clip-rule="evenodd" class="header__links__item__image">
              <path :d="icon(link)" style="pointer-events: none"></path>
            </svg>
          </a>
        </div>
      </div>
      <div class="menu">
        <div class="menu__item" v-for="item in menu" :key="item.title">
          <div class="menu__item__title">{{item.h1}}</div>
          <div v-for="child in item.children" :key="child.h1" class="menu__item__item">
            <router-link v-if="hasRouter && !isExternal(child.href)" tag="a" :to="child.href">{{child.h1}}</router-link>
            <a v-else :href="child.href" target="_blank">{{child.h1}}</a>
          </div>
        </div>
      </div>
    </div>
    <div class="smallprint" v-if="smallprint">{{smallprint}}</div>
  </div>
</template>

<style scoped>
a {
  color: inherit;
  text-decoration: none;
}

.component__footer__container {
  font-family: var(--ds-font-family, inherit);
  background-color: var(--grey-14, rgb(21, 24, 49));
  color: var(--white-100, white);
  padding-top: 2rem;
  padding-bottom: 2rem;
  overflow-x: hidden;
}

.grid {
  display: grid;
  grid-template-columns: 1fr 2fr;
  margin-bottom: 2rem;
}

.header {
  display: grid;
  gap: 1rem;
  grid-auto-flow: row;
  align-content: space-between;
  justify-content: space-between;
}

.header__title {
  font-size: 1.25rem;
  margin-bottom: 2rem;
}

.header__links {
  display: flex;
  flex-direction: row;
  flex-wrap: wrap;
  position: relative;
}

.header__links__popover {
  position: absolute;
  top: -3.5rem;
  transform: translate(var(--pos-x), var(--pos-y));
  pointer-events: none;
  transition: all .25s;
}

.header__links__popover__content {
  background: var(--white-100, white);
  color: var(--black, black);
  position: absolute;
  white-space: nowrap;
  transform: translateX(-50%);
  font-size: .8125rem;
  padding: 7px 12px;
  border-radius: .25rem;
}

.header__links__popover__content:after {
  content: "";
  position: absolute;
  width: 8px;
  height: 8px;
  display: block;
  background-color: var(--white-100, white);
  mask-image: url("data:image/svg+xml, <svg xmlns='http://www.w3.org/2000/svg' width='100%' height='100%' viewBox='0 0 24 24'><path d='M12 21l-12-18h24z'/></svg>");
  background-repeat: no-repeat;
  top: 100%;
  left: 50%;
  transform: translateX(-50%) translateY(-13%) scaleX(2);
}

.header__links__item {
  fill: var(--white-100, white);
  margin-right: 0.75rem;
  margin-bottom: 0.75rem;
  display: block;
  height: 24px;
  width: 24px;
}

.header__links__item__image {
  position: block;
}

.menu {
  display: grid;
  gap: 1.5rem;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
}

.menu__item__title {
  margin-bottom: 1rem;
  font-weight: var(--ds-h6-font-weight, 800);
  font-size: var(--ds-h6-font-size, 0.75rem);
  text-transform: var(--ds-h6-text-transform, uppercase);
  letter-spacing: var(--ds-h6-letter-spacing, 0.2em);
}

.menu__item__item {
  padding-top: 0.5rem;
  padding-bottom: 0.5rem;
  display: block;
  font-size: var(--ds-p2-font-size, 0.875rem);
  line-height: var(--ds-p2-line-height, 1.25);
}

.smallprint {
  font-size: var(--ds-p3-font-size, 0.825rem);
}

.fade-active-enter {
  transition: all .25s;
}

.fade-enter {
  opacity: 0;
}

.fade-enter-to {
  opacity: 1;
}

.fade-active-leave {
  transition: all .25s;
}

.fade-leave {
  opacity: 1;
}

.fade-leave-to {
  opacity: 0;
}

@media screen and (max-width: 800px) {
  .grid {
    display: block;
  }

  .header {
    grid-auto-flow: column;
    margin-bottom: 3rem;
  }
}

@media screen and (max-width: 400px) {
  .header {
    display: block;
  }
}
</style>

<script>
const iconList = [
  [
    "blog.cosmos.network",
    "M24 24h-24v-24h24v24zm-4.03-5.649v-.269l-1.247-1.224c-.11-.084-.165-.222-.142-.359v-8.998c-.023-.137.032-.275.142-.359l1.277-1.224v-.269h-4.422l-3.152 7.863-3.586-7.863h-4.638v.269l1.494 1.799c.146.133.221.327.201.523v7.072c.044.255-.037.516-.216.702l-1.681 2.038v.269h4.766v-.269l-1.681-2.038c-.181-.186-.266-.445-.232-.702v-6.116l4.183 9.125h.486l3.593-9.125v7.273c0 .194 0 .232-.127.359l-1.292 1.254v.269h6.274z"
  ],
  [
    "twitter",
    "M24 4.557c-.883.392-1.832.656-2.828.775 1.017-.609 1.798-1.574 2.165-2.724-.951.564-2.005.974-3.127 1.195-.897-.957-2.178-1.555-3.594-1.555-3.179 0-5.515 2.966-4.797 6.045-4.091-.205-7.719-2.165-10.148-5.144-1.29 2.213-.669 5.108 1.523 6.574-.806-.026-1.566-.247-2.229-.616-.054 2.281 1.581 4.415 3.949 4.89-.693.188-1.452.232-2.224.084.626 1.956 2.444 3.379 4.6 3.419-2.07 1.623-4.678 2.348-7.29 2.04 2.179 1.397 4.768 2.212 7.548 2.212 9.142 0 14.307-7.721 13.995-14.646.962-.695 1.797-1.562 2.457-2.549z"
  ],
  [
    "linkedin",
    "M0 0v24h24v-24h-24zm8 19h-3v-11h3v11zm-1.5-12.268c-.966 0-1.75-.79-1.75-1.764s.784-1.764 1.75-1.764 1.75.79 1.75 1.764-.783 1.764-1.75 1.764zm13.5 12.268h-3v-5.604c0-3.368-4-3.113-4 0v5.604h-3v-11h3v1.765c1.397-2.586 7-2.777 7 2.476v6.759z"
  ],
  [
    "reddit",
    "M14.238 15.348c.085.084.085.221 0 .306-.465.462-1.194.687-2.231.687l-.008-.002-.008.002c-1.036 0-1.766-.225-2.231-.688-.085-.084-.085-.221 0-.305.084-.084.222-.084.307 0 .379.377 1.008.561 1.924.561l.008.002.008-.002c.915 0 1.544-.184 1.924-.561.085-.084.223-.084.307 0zm-3.44-2.418c0-.507-.414-.919-.922-.919-.509 0-.923.412-.923.919 0 .506.414.918.923.918.508.001.922-.411.922-.918zm13.202-.93c0 6.627-5.373 12-12 12s-12-5.373-12-12 5.373-12 12-12 12 5.373 12 12zm-5-.129c0-.851-.695-1.543-1.55-1.543-.417 0-.795.167-1.074.435-1.056-.695-2.485-1.137-4.066-1.194l.865-2.724 2.343.549-.003.034c0 .696.569 1.262 1.268 1.262.699 0 1.267-.566 1.267-1.262s-.568-1.262-1.267-1.262c-.537 0-.994.335-1.179.804l-2.525-.592c-.11-.027-.223.037-.257.145l-.965 3.038c-1.656.02-3.155.466-4.258 1.181-.277-.255-.644-.415-1.05-.415-.854.001-1.549.693-1.549 1.544 0 .566.311 1.056.768 1.325-.03.164-.05.331-.05.5 0 2.281 2.805 4.137 6.253 4.137s6.253-1.856 6.253-4.137c0-.16-.017-.317-.044-.472.486-.261.82-.766.82-1.353zm-4.872.141c-.509 0-.922.412-.922.919 0 .506.414.918.922.918s.922-.412.922-.918c0-.507-.413-.919-.922-.919z"
  ],
  [
    "t.me",
    "M12,0c-6.626,0 -12,5.372 -12,12c0,6.627 5.374,12 12,12c6.627,0 12,-5.373 12,-12c0,-6.628 -5.373,-12 -12,-12Zm3.224,17.871c0.188,0.133 0.43,0.166 0.646,0.085c0.215,-0.082 0.374,-0.267 0.422,-0.491c0.507,-2.382 1.737,-8.412 2.198,-10.578c0.035,-0.164 -0.023,-0.334 -0.151,-0.443c-0.129,-0.109 -0.307,-0.14 -0.465,-0.082c-2.446,0.906 -9.979,3.732 -13.058,4.871c-0.195,0.073 -0.322,0.26 -0.316,0.467c0.007,0.206 0.146,0.385 0.346,0.445c1.381,0.413 3.193,0.988 3.193,0.988c0,0 0.847,2.558 1.288,3.858c0.056,0.164 0.184,0.292 0.352,0.336c0.169,0.044 0.348,-0.002 0.474,-0.121c0.709,-0.669 1.805,-1.704 1.805,-1.704c0,0 2.084,1.527 3.266,2.369Zm-6.423,-5.062l0.98,3.231l0.218,-2.046c0,0 3.783,-3.413 5.941,-5.358c0.063,-0.057 0.071,-0.153 0.019,-0.22c-0.052,-0.067 -0.148,-0.083 -0.219,-0.037c-2.5,1.596 -6.939,4.43 -6.939,4.43Z"
  ],
  [
    "discord.gg",
    "M19.54 0c1.356 0 2.46 1.104 2.46 2.472v21.528l-2.58-2.28-1.452-1.344-1.536-1.428.636 2.22h-13.608c-1.356 0-2.46-1.104-2.46-2.472v-16.224c0-1.368 1.104-2.472 2.46-2.472h16.08zm-4.632 15.672c2.652-.084 3.672-1.824 3.672-1.824 0-3.864-1.728-6.996-1.728-6.996-1.728-1.296-3.372-1.26-3.372-1.26l-.168.192c2.04.624 2.988 1.524 2.988 1.524-1.248-.684-2.472-1.02-3.612-1.152-.864-.096-1.692-.072-2.424.024l-.204.024c-.42.036-1.44.192-2.724.756-.444.204-.708.348-.708.348s.996-.948 3.156-1.572l-.12-.144s-1.644-.036-3.372 1.26c0 0-1.728 3.132-1.728 6.996 0 0 1.008 1.74 3.66 1.824 0 0 .444-.54.804-.996-1.524-.456-2.1-1.416-2.1-1.416l.336.204.048.036.047.027.014.006.047.027c.3.168.6.3.876.408.492.192 1.08.384 1.764.516.9.168 1.956.228 3.108.012.564-.096 1.14-.264 1.74-.516.42-.156.888-.384 1.38-.708 0 0-.6.984-2.172 1.428.36.456.792.972.792.972zm-5.58-5.604c-.684 0-1.224.6-1.224 1.332 0 .732.552 1.332 1.224 1.332.684 0 1.224-.6 1.224-1.332.012-.732-.54-1.332-1.224-1.332zm4.38 0c-.684 0-1.224.6-1.224 1.332 0 .732.552 1.332 1.224 1.332.684 0 1.224-.6 1.224-1.332 0-.732-.54-1.332-1.224-1.332z"
  ],
  [
    "youtube",
    "M19.615 3.184c-3.604-.246-11.631-.245-15.23 0-3.897.266-4.356 2.62-4.385 8.816.029 6.185.484 8.549 4.385 8.816 3.6.245 11.626.246 15.23 0 3.897-.266 4.356-2.62 4.385-8.816-.029-6.185-.484-8.549-4.385-8.816zm-10.615 12.816v-8l8 3.993-8 4.007z"
  ]
];

const iconUnknown =
  "M13.144 8.171c-.035-.066.342-.102.409-.102.074.009-.196.452-.409.102zm-2.152-3.072l.108-.031c.064.055-.072.095-.051.136.086.155.021.248.008.332-.014.085-.104.048-.149.093-.053.066.258.075.262.085.011.033-.375.089-.304.171.096.136.824-.195.708-.176.225-.113.029-.125-.097-.19-.043-.215-.079-.547-.213-.68l.088-.102c-.206-.299-.36.362-.36.362zm13.008 6.901c0 6.627-5.373 12-12 12-6.628 0-12-5.373-12-12s5.372-12 12-12c6.627 0 12 5.373 12 12zm-8.31-5.371c-.006-.146-.19-.284-.382-.031-.135.174-.111.439-.184.557-.104.175.567.339.567.174.025-.277.732-.063.87-.025.248.069.643-.226.211-.381-.355-.13-.542-.269-.574-.523 0 0 .188-.176.106-.166-.218.027-.614.786-.614.395zm6.296 5.371c0-1.035-.177-2.08-.357-2.632-.058-.174-.189-.312-.359-.378-.256-.1-1.337.597-1.5.254-.107-.229-.324.146-.572.008-.12-.066-.454-.515-.605-.46-.309.111.474.964.688 1.076.201-.152.852-.465.992-.038.268.804-.737 1.685-1.251 2.149-.768.694-.624-.449-1.147-.852-.275-.211-.272-.66-.55-.815-.124-.07-.693-.725-.688-.813l-.017.166c-.094.071-.294-.268-.315-.321 0 .295.48.765.639 1.001.271.405.416.995.748 1.326.178.178.858.914 1.035.898.193-.017.803-.458.911-.433.644.152-1.516 3.205-1.721 3.583-.169.317.138 1.101.113 1.476-.029.433-.37.573-.693.809-.346.253-.265.745-.556.925-.517.318-.889 1.353-1.623 1.348-.216-.001-1.14.36-1.261.007-.094-.256-.22-.45-.353-.703-.13-.248-.015-.505-.173-.724-.109-.152-.475-.497-.508-.677-.002-.155.117-.626.28-.708.229-.117.044-.458.016-.656-.048-.354-.267-.646-.53-.851-.389-.299-.188-.537-.097-.964 0-.204-.124-.472-.398-.392-.564.164-.393-.44-.804-.413-.296.021-.538.209-.813.292-.346.104-.7-.082-1.042-.125-1.407-.178-1.866-1.786-1.499-2.946.037-.19-.114-.542-.048-.689.158-.352.48-.747.762-1.014.158-.15.361-.112.547-.229.287-.181.291-.553.572-.781.4-.325.946-.318 1.468-.388.278-.037 1.336-.266 1.503-.06 0 .038.191.604-.019.572.433.023 1.05.749 1.461.579.211-.088.134-.736.567-.423.262.188 1.436.272 1.68.069.15-.124.234-.93.052-1.021.116.115-.611.124-.679.098-.12-.044-.232.114-.425.025.116.055-.646-.354-.218-.667-.179.131-.346-.037-.539.107-.133.108.062.18-.128.274-.302.153-.53-.525-.644-.602-.116-.076-1.014-.706-.77-.295l.789.785c-.039.025-.207-.286-.207-.059.053-.135.02.579-.104.347-.055-.089.09-.139.006-.268 0-.085-.228-.168-.272-.226-.125-.155-.457-.497-.637-.579-.05-.023-.764.087-.824.11-.07.098-.13.201-.179.311-.148.055-.287.126-.419.214l-.157.353c-.068.061-.765.291-.769.3.029-.075-.487-.171-.453-.321.038-.165.213-.68.168-.868-.048-.197 1.074.284 1.146-.235.029-.225.046-.487-.313-.525.068.008.695-.246.799-.36.146-.168.481-.442.724-.442.284 0 .223-.413.354-.615.131.053-.07.376.087.507-.01-.103.445.057.489.033.104-.054.684-.022.594-.294-.1-.277.051-.195.181-.253-.022.009.34-.619.402-.413-.043-.212-.421.074-.553.063-.305-.024-.176-.52-.061-.665.089-.115-.243-.256-.247-.036-.006.329-.312.627-.241 1.064.108.659-.735-.159-.809-.114-.28.17-.509-.214-.364-.444.148-.235.505-.224.652-.476.104-.178.225-.385.385-.52.535-.449.683-.09 1.216-.041.521.048.176.124.104.324-.069.19.286.258.409.099.07-.092.229-.323.298-.494.089-.222.901-.197.334-.536-.374-.223-2.004-.672-3.096-.672-.236 0-.401.263-.581.412-.356.295-1.268.874-1.775.698-.519-.179-1.63.66-1.808.666-.065.004.004-.634.358-.681-.153.023 1.247-.707 1.209-.859-.046-.18-2.799.822-2.676 1.023.059.092.299.092-.016.294-.18.109-.372.801-.541.801-.505.221-.537-.435-1.099.409l-.894.36c-1.328 1.411-2.247 3.198-2.58 5.183-.013.079.334.226.379.28.112.134.112.712.167.901.138.478.479.744.74 1.179.154.259.41.914.329 1.186.108-.178 1.07.815 1.246 1.022.414.487.733 1.077.061 1.559-.217.156.33 1.129.048 1.368l-.361.093c-.356.219-.195.756.021.982 1.818 1.901 4.38 3.087 7.22 3.087 5.517 0 9.989-4.472 9.989-9.989zm-11.507-6.357c.125-.055.293-.053.311-.22.015-.148.044-.046.08-.1.035-.053-.067-.138-.11-.146-.064-.014-.108.069-.149.104l-.072.019-.068.087.008.048-.087.106c-.085.084.002.139.087.102z";

/**
 * The `Footer` component is used to display information,
 * such as social media links and links to different parts
 * of the website, at the bottom of a page. `Footer` takes
 * the full width of the page, a parent container can be used
 * to center the footer and set `max-width`.
 *
 * The following CSS custom variables are used to theme the
 * component:
 *
 * `--ds-font-family`
 *
 * `--grey-14`
 *
 * `--white-100` — main text color.
 *
 * `--black` — text color in popups.
 *
 * `--ds-h6-*` — menu subsection headings.
 *
 * `--ds-h6-font-weight`
 *
 * `--ds-h6-font-size`
 *
 * `--ds-h6-text-transform`
 *
 * `--ds-h6-letter-spacing`
 *
 * `--ds-p2-font-size` — menu subsection items.
 *
 * `--ds-p3-font-size` — text at the bottom of the component.
 */
export default {
  props: {
    /**
     * Title in the top left corner of the component.
     */
    h1: {
      type: String,
      default: ""
    },
    /**
     * An element of the array can be a `String`, in which case the string
     * is used as a URL for the link and default icon is used.
     * An element can also be an object with the following properties:
     * `title`, `href`, `icon`. `title` is used for the value inside the tooltip,
     * `href` is the URL and `icon` specifies an icon (currently, from a set of 6).
     */
    links: {
      type: Array,
      default: () => []
    },
    /**
     * An element of the array is an object `{h1: String, children: Array}`.
     * `h1` is a title of a subsection. Element of a `children` array is
     * an object `{h1: String, href: String}`. `h1` is the text of a link
     * and `href` is the link`s URL.
     */
    menu: {
      type: Array,
      default: () => []
    },
    /**
     * Text that appears at the bottom of the component.
     */
    smallprint: {
      type: String,
      default: ""
    }
  },
  data: function() {
    return {
      popover: null,
      popoverActive: null,
      popoverTimer: null,
      popoverX: null,
      popoverY: null
    }
  },
  computed: {
    hasRouter() {
      return this.$router
    }
  },
  methods: {
    linkMouseover(link, e, entering) {
      const leaving = !entering
      const parent = this.$refs.links.getBoundingClientRect()
      const element = e.target.getBoundingClientRect()
      if (entering) {
        this.popoverX = element.x - parent.x + element.width / 2
        this.popoverY = element.y - parent.y + element.height / 2
        this.popover = link
        clearTimeout(this.popoverTimer)
      }
      if (leaving) {
        this.popoverTimer = setTimeout(() => {
          this.popover = false
        }, 500)
      }
    },
    isExternal(url) {
      const match = url.match(/^([^:\/?#]+:)?(?:\/\/([^\/?#]*))?([^?#]+)?(\?[^#]*)?(#.*)?/);
      if (typeof match[1] === "string" && match[1].length > 0 && match[1].toLowerCase() !== location.protocol) return true;
      if (typeof match[2] === "string" && match[2].length > 0 && match[2].replace(new RegExp(":("+{"http:":80,"https:":443}[location.protocol]+")?$"), "") !== location.host) return true;
      return false;
    },
    url(link) {
      return link.url || link
    },
    icon(link) {
      let iconPath;
      let url = link.url || link
      iconList.forEach(icon => {
        if (link.icon && link.icon.match(icon[0])) {
          iconPath = icon[1]
        } else if (url.match(icon[0])) {
          iconPath = icon[1];
        }
      });
      return iconPath || iconUnknown;
    }
  }
};
</script>
