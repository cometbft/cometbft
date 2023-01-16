<template>
  <div v-if="visited">
    <transition name="fade" appear>
      <div class="banner-wrapper" v-if="show">
        <div class="wrapper">
          <div class="message">By using this website, you agree to our <a href="https://www.cookiesandyou.com" target="_blank" rel="noopener" style="color: #505FFF;">Cookie Policy</a>.</div>
            <div class="box" @click="close">
              <span class="icon-cross">
              <svg
                width="16"
                height="16"
                viewBox="0 0 14 14"
                fill="none"
                xmlns="http://www.w3.org/2000/svg"
              >
                <path
                  d="M1.66669 1.66669L12.3334 12.3334M12.3334 1.66669L1.66664 12.3334"
                  stroke="#A2A3AD"
                  stroke-width="1.5"
                  stroke-linecap="round"
                />
              </svg>
            </span>
          </div>
        </div>
      </div>
    </transition>
  </div>
</template>

<script>
import * as Cookie from 'tiny-cookie'

/**
 * Displays a banner with a link to cookie policy.
 * Closing the banner sets `cookie-consent-accepted`
 * (or specified by `storageName` prop)
 * cookie to `true`. When this cookie is set, the
 * banner is not displayed.
 */
export default {
  props: {
    /**
     * Name of the cookie that indicates that a user
     * has accepted use of cookies on the website.
     */
    storageName: {
      type: String,
      default: 'cookie-consent-accepted'
    },
    /**
     * Object passed to `setCookieStatus` function
     * of `tiny-cookie` to customize the properties
     * of the cookie.
     */
    cookieOptions: {
      type: Object,
      default: () => ({ expires: '1M' })
    },
  },
  data: function() {
    return {
      show: true,
      visited: true
    };
  },
  mounted () {
    this.init()
  },
  methods: {
    init() {
      this.checkCookieStatus();
    },
    checkCookieStatus() {
      if (Cookie.get(this.storageName, false)) {
        this.visited = false
      } else this.visited = true;
    },
    setCookieStatus() {
      Cookie.set(this.storageName, true, this.cookieOptions);
    },
    close() {
      this.show = false;
      this.setCookieStatus();
    },
  }
};
</script>

<style scoped>
.banner-wrapper {
  box-sizing: border-box;
  box-shadow: 0px 24px 40px rgba(22, 25, 49, 0.1),
    0px 10px 16px rgba(22, 25, 49, 0.08), 0px 1px 0px rgba(22, 25, 49, 0.05);
  position: fixed;
  bottom: 2rem;
  left: 50%;
  transform: translateX(-50%);
  z-index: 3000;
  background: var(--white-100, white);
  border-width: 0.0625rem;
  border-style: solid;
  border-color: transparent;
  border-image: initial;
  border-radius: 0.5rem;
}
.wrapper {
  display: flex;
}
.message {
  font-size: 0.8125rem;
  line-height: 1.125rem;
  letter-spacing: 0.01em;
  align-items: stretch;
  justify-content: center;
  padding: 0.8125rem 0 0.8125rem 0.8125rem;
  position: relative;
}
.box {
  display: flex;
  justify-content: center;
  align-items: center;
  width: 3.5rem;
  flex: 0 0 3.5rem;
  flex-direction: column;
  padding: 0;
  position: relative;
  cursor: pointer;
}
.icon-cross {
  display: flex;
  justify-content: center;
  align-items: center;
}
.box:hover .icon-cross {
  height: 2rem;
  width: 2rem;
  background-color: var(--grey-95, #eff0f5);
  border-radius: 1rem;
  transition: background-color 0.2s;
}
.box:hover path {
  stroke: var(--grey-44, #686a79);
  transition: stroke 0.2s;
}

.fade-enter-active {
  transition: opacity 0.2s ease-out, transform 0.5s cubic-bezier(0.22, 1, 0.36, 1);
}
.fade-leave-active {
  transition: opacity 0.3s cubic-bezier(0.32, 0, 0.67, 0), transform 0.3s cubic-bezier(0.32, 0, 0.67, 0);
}
.fade-enter,
.fade-leave-to {
  opacity: 0;
  transform: translate(-50%, 100%);
}
@media screen and (max-width: 435px) {
  .banner-wrapper {
    border-radius: 0.5rem;
    bottom: 1rem;
    white-space: normal;
    width: 90%;
  }
  .message {
    font-size: 0.90rem;
    line-height: 1.125rem;
  }
}
</style>
