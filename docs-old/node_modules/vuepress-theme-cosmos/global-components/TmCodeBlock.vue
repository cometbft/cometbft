<template>
  <span>
    <span
      class="container"
      ref="container"
      :class="[
        `codeblock__hasfooter__${!!url}`,
        `codeblock__is-expandable__${!!isExpandable}`,
        `codeblock__expanded__${!!expanded}`,
      ]"
    >
      <span class="body__container">
        <span class="body__block">
          <span class="icons">
            <span
              class="icons__item"
              v-if="height &gt; 300 &amp;&amp; expanded"
            >
              <svg
                class="icons__item__icon"
                width="24"
                height="24"
                viewBox="0 0 24 24"
                fill="none"
                xmlns="http://www.w3.org/2000/svg"
                @click="expand(false)"
              >
                <path
                  fill-rule="evenodd"
                  clip-rule="evenodd"
                  d="M12.5303 10.7803L12 11.3107L11.4697 10.7803L6.96967 6.28033C6.67678 5.98744 6.67678 5.51256 6.96967 5.21967C7.26256 4.92678 7.73744 4.92678 8.03033 5.21967L11.25 8.43934L11.25 1.5C11.25 1.08579 11.5858 0.75 12 0.75C12.4142 0.75 12.75 1.08579 12.75 1.5L12.75 8.43934L15.9697 5.21967C16.2626 4.92678 16.7374 4.92678 17.0303 5.21967C17.3232 5.51256 17.3232 5.98744 17.0303 6.28033L12.5303 10.7803ZM12.5303 13.2197L12 12.6893L11.4697 13.2197L6.96967 17.7197C6.67678 18.0126 6.67678 18.4874 6.96967 18.7803C7.26256 19.0732 7.73744 19.0732 8.03033 18.7803L11.25 15.5607L11.25 22.5C11.25 22.9142 11.5858 23.25 12 23.25C12.4142 23.25 12.75 22.9142 12.75 22.5L12.75 15.5607L15.9697 18.7803C16.2626 19.0732 16.7374 19.0732 17.0303 18.7803C17.3232 18.4874 17.3232 18.0126 17.0303 17.7197L12.5303 13.2197Z"
                ></path>
              </svg>
              <span class="icons__item__tooltip">
                Collapse
              </span>
            </span>
            <span class="icons__item">
              <svg
                class="icons__item__icon"
                width="24"
                height="24"
                viewBox="0 0 24 24"
                xmlns="http://www.w3.org/2000/svg"
                @click="copy(source)"
              >
                <path
                  fill-rule="evenodd"
                  clip-rule="evenodd"
                  d="M11 0.25C10.0335 0.25 9.25 1.0335 9.25 2V4.5H10.75V2C10.75 1.86193 10.8619 1.75 11 1.75H21C21.1381 1.75 21.25 1.86193 21.25 2V16C21.25 16.1381 21.1381 16.25 21 16.25H16.5V17.75H21C21.9665 17.75 22.75 16.9665 22.75 16V2C22.75 1.0335 21.9665 0.25 21 0.25H11ZM3 6.25C2.0335 6.25 1.25 7.0335 1.25 8V22C1.25 22.9665 2.0335 23.75 3 23.75H13C13.9665 23.75 14.75 22.9665 14.75 22V8C14.75 7.0335 13.9665 6.25 13 6.25H3ZM2.75 8C2.75 7.86193 2.86193 7.75 3 7.75H13C13.1381 7.75 13.25 7.86193 13.25 8V22C13.25 22.1381 13.1381 22.25 13 22.25H3C2.86193 22.25 2.75 22.1381 2.75 22V8Z"
                ></path>
              </svg>
              <span class="icons__item__tooltip">
                {{ copied ? "Copied!" : "Copy" }}
              </span>
            </span>
          </span>
          <span class="body" :style="{ '--max-height': maxHeight }" ref="body">
            <span class="body__wrapper">
              <span class="body__code" v-html="highlighted(source)"></span>
            </span>
          </span>
        </span>
        <span class="expand" v-if="isExpandable">
          <span
            class="expand__item expand__item__expand"
            @click="expand(true)"
            v-if="!expanded"
          >
            <span>Expand</span>
            <svg
              class="expand__item__icon"
              width="100%"
              height="100%"
              viewBox="0 0 16 16"
              fill="none"
              xmlns="http://www.w3.org/2000/svg"
            >
              <path
                d="M7.25 0.99998C7.25 0.585766 7.58578 0.24998 8 0.24998C8.41421 0.24998 8.75 0.585766 8.75 0.99998L7.25 0.99998ZM8 14.8333L8.53033 15.3636L8 15.894L7.46967 15.3636L8 14.8333ZM2.46967 10.3636C2.17678 10.0708 2.17678 9.59588 2.46967 9.30298C2.76256 9.01009 3.23744 9.01009 3.53033 9.30298L2.46967 10.3636ZM12.4697 9.30298C12.7626 9.01009 13.2374 9.01009 13.5303 9.30298C13.8232 9.59587 13.8232 10.0707 13.5303 10.3636L12.4697 9.30298ZM8.75 0.99998L8.75 14.8333L7.25 14.8333L7.25 0.99998L8.75 0.99998ZM7.46967 15.3636L2.46967 10.3636L3.53033 9.30298L8.53033 14.303L7.46967 15.3636ZM13.5303 10.3636L8.53033 15.3636L7.46967 14.303L12.4697 9.30298L13.5303 10.3636Z"
                fill="black"
              ></path>
            </svg>
          </span>
          <span
            class="expand__item expand__item__collapse"
            @click="expand(false, true)"
            v-if="height &gt; 300 &amp;&amp; expanded"
          >
            <svg
              width="100%"
              height="100%"
              viewBox="0 0 12 24"
              fill="none"
              xmlns="http://www.w3.org/2000/svg"
            >
              <path
                fill-rule="evenodd"
                clip-rule="evenodd"
                d="M6.53033 10.7803L6 11.3107L5.46967 10.7803L0.96967 6.28033C0.676777 5.98744 0.676777 5.51256 0.96967 5.21967C1.26256 4.92678 1.73744 4.92678 2.03033 5.21967L5.25 8.43934L5.25 1.5C5.25 1.08579 5.58578 0.75 6 0.75C6.41421 0.75 6.75 1.08579 6.75 1.5L6.75 8.43934L9.96967 5.21967C10.2626 4.92678 10.7374 4.92678 11.0303 5.21967C11.3232 5.51256 11.3232 5.98744 11.0303 6.28033L6.53033 10.7803ZM6.53033 13.2197L6 12.6893L5.46967 13.2197L0.96967 17.7197C0.676777 18.0126 0.676777 18.4874 0.96967 18.7803C1.26256 19.0732 1.73744 19.0732 2.03033 18.7803L5.25 15.5607L5.25 22.5C5.25 22.9142 5.58578 23.25 6 23.25C6.41421 23.25 6.75 22.9142 6.75 22.5L6.75 15.5607L9.96967 18.7803C10.2626 19.0732 10.7374 19.0732 11.0303 18.7803C11.3232 18.4874 11.3232 18.0126 11.0303 17.7197L6.53033 13.2197Z"
                fill="#2E3148"
              ></path>
            </svg>
          </span>
        </span>
      </span>
      <span class="footer" v-if="url">
        <span class="footer__filename">
          {{ filename(url) }}
        </span>
        <a
          class="footer__source"
          :href="url"
          target="_blank"
          rel="noreferrer noopener"
        >
          <span>View source</span>
          <svg
            class="footer__source__icon"
            width="16"
            height="16"
            viewBox="0 0 16 16"
            xmlns="http://www.w3.org/2000/svg"
          >
            <path
              d="M5 2.5L10.5 8L5 13.5"
              stroke-width="1.5"
              stroke-linecap="round"
            ></path>
          </svg>
        </a>
      </span>
    </span>
  </span>
</template>

<style scoped>
a {
  text-decoration: none;
}

.body__code {
  white-space: pre;
}
span {
  display: block;
}
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s;
}
.fade-enter,
.fade-leave-to {
  opacity: 0;
}
.fade-enter-to,
.fade-leave {
  opacity: 1;
}
.container {
  border-radius: 0.5rem;
  background: #2e3148;
}
.body__container {
  position: relative;
}
.body__container:hover .icons {
  opacity: 1;
}
.body {
  color: rgba(255, 255, 255, 0.8);
  overflow-x: scroll;
  padding-left: 1rem;
  padding-right: 1rem;
  padding-top: 1.375rem;
  padding-bottom: 1rem;
  overflow-y: hidden;
  position: relative;
  line-height: 1.75;
  scrollbar-color: rgba(255, 255, 255, 0.2) rgba(255, 255, 255, 0.1);
  scrollbar-width: thin;
  font-size: 0.8125rem;
  line-height: 1.25rem;
}

.body::-webkit-scrollbar {
  background: rgba(255, 255, 255, 0.1);
  height: 6px;
}

.body::-webkit-scrollbar-thumb {
  background: rgba(255, 255, 255, 0.2);
  border-radius: 6px;
}

.codeblock__is-expandable__true .body {
  max-height: 700px;
}

.codeblock__is-expandable__true.codeblock__expanded__true .body {
  max-height: var(--max-height);
}
.body__wrapper {
  font-family: "JetBrains Mono", "Menlo", "Monaco", monospace;
  -webkit-font-feature-settings: "liga" on, "calt" on;
  -webkit-font-smoothing: antialiased;
  text-rendering: optimizeLegibility;
  font-size: 0.8125rem;
  display: inline-block;
  line-height: 1.25rem;
}
.body.body__hasfooter__true {
  border-bottom-left-radius: 0;
  border-bottom-right-radius: 0;
}
.expand {
  padding-top: 1.5rem;
  padding-bottom: 1.5rem;
  position: absolute;
  bottom: 0;
  display: grid;
  grid-template-columns: 1fr 1fr 1fr;
  width: 100%;
  color: #161931;
  padding-right: 1.5rem;
  padding-left: 1.5rem;
  font-family: var(--ds-font-family, inherit);
  box-sizing: border-box;
}
.codeblock__expanded__false .expand {
  background: linear-gradient(180deg, rgba(22, 25, 49, 0) 0%, #161931 100%);
}
.codeblock__hasfooter__false.codeblock__expanded__false .expand {
  border-bottom-left-radius: 0.5rem;
  border-bottom-right-radius: 0.5rem;
}
.expand__item {
  text-transform: uppercase;
  background-color: #dadce6;
  display: grid;
  justify-self: center;
  grid-auto-flow: column;
  gap: 0.5rem;
  align-items: center;
  font-weight: 500;
  padding: 0.5rem 1rem;
  line-height: 1;
  letter-spacing: 0.02em;
  font-size: 0.8125rem;
  height: 2rem;
  cursor: pointer;
  border-radius: 1000px;
  box-shadow: 0px 16px 32px rgba(22, 25, 49, 0.08),
    0px 8px 12px rgba(22, 25, 49, 0.06), 0px 1px 0px rgba(22, 25, 49, 0.05);
  box-sizing: border-box;
}
.expand__item__expand {
  grid-area: 1/2/1/3;
  justify-self: center;
}
.expand__item__collapse {
  grid-area: 1/3/1/4;
  justify-self: flex-end;
  height: 2rem;
  max-width: 3rem;
  padding-top: 0.3rem;
  padding-bottom: 0.3rem;
}
.expand__item__icon {
  height: 1em;
  width: auto;
}
.icons {
  transition: all 0.1s;
  position: absolute;
  top: 0;
  right: 0;
  padding: 0.5rem;
  opacity: 0;
  display: flex;
  z-index: 100;
}
.icons__item {
  margin-left: 0.5rem;
  cursor: pointer;
  border-radius: 0.25rem;
  position: relative;
  background: rgba(46, 49, 72, 0.7);
}
.icons__item:active .icons__item__icon {
  fill: #66a1ff;
}
.icons__item:hover .icons__item__tooltip {
  opacity: 1;
}
.icons__item:hover .icons__item__tooltip:hover {
  opacity: 0;
}
.icons__item__tooltip {
  font-family: var(--ds-font-family, inherit);
  color: #fff;
  position: absolute;
  top: -2.05rem;
  left: 50%;
  transform: translateX(-50%);
  background: #161931;
  font-size: 0.8125rem;
  opacity: 0;
  border-radius: 0.25rem;
  padding: 0.5rem 0.75rem;
  transition: all 0.25s 0.5s;
}
.icons__item__tooltip:before {
  content: "";
  position: absolute;
  width: 8px;
  height: 8px;
  display: block;
  background-color: #161931;
  mask-image: url("data:image/svg+xml, <svg xmlns='http://www.w3.org/2000/svg' width='100%' height='100%' viewBox='0 0 24 24'><path d='M12 21l-12-18h24z'/></svg>");
  background-repeat: no-repeat;
  top: 100%;
  left: 50%;
  transform: translateX(-50%) translateY(-13%) scaleX(2);
  position: absolute;
  font-size: 1rem;
  left: 50%;
}
.icons__item__icon {
  fill: #fff;
  padding: 0.75rem;
  display: block;
}
.icons__item:hover {
  fill: #66a1ff;
  background: #43465a;
}
.footer {
  background-color: #161931;
  color: #fff;
  display: flex;
  justify-content: space-between;
  padding: 0.75rem 1rem;
  border-bottom-left-radius: 0.5rem;
  border-bottom-right-radius: 0.5rem;
  font-size: 0.8125rem;
  line-height: 1;
  font-family: var(--ds-font-family, inherit);
}
.footer__source {
  color: #66a1ff;
  font-weight: 500;
  stroke: #66a1ff;
  align-items: center;
  display: flex;
  box-shadow: none;
  outline: none;
}
.footer__source:after {
  display: none;
}
.footer__source:visited {
  color: #66a1ff;
}
.footer__source:visited:hover {
  color: #66a1ff;
}
.footer__source:hover {
  box-shadow: none;
  color: #66a1ff;
}
.footer__source:active {
  color: #66a1ff;
}
.footer__source__icon {
  margin-left: 0.5rem;
}
::v-deep .token.keyword {
  color: #c678dd;
}
::v-deep .token.comment {
  opacity: 0.5;
}
::v-deep .token.function {
  color: #61afef;
}
::v-deep .token.builtin {
  color: #e06c75;
}
::v-deep .token.string {
  color: #98c379;
}
::v-deep .token.operator {
  color: #56b6c2;
}
::v-deep .token.boolean {
  color: #d19a66;
}
</style>

<script>
import Prism from "prismjs";
import "prismjs/components/prism-go.min.js";
import "prismjs/components/prism-rust.min.js";
import "prismjs/components/prism-markdown.min.js";
import "prismjs/components/prism-bash.min.js";
import "prismjs/components/prism-json.min.js";
import "prismjs/components/prism-protobuf.min.js";
import "prismjs/components/prism-solidity.min.js";
import "prismjs/components/prism-python.min.js";
import copy from "clipboard-copy";
import { Base64 } from "js-base64";

export default {
  props: {
    /**
     * Code rendered in the body of the block
     */
    value: {
      type: String,
    },
    /**
     * Code rendered in the body of the block in base64
     */
    base64: {
      type: String,
    },
    /**
     * URL for "View source" link and filename in the code-block's footer
     */
    url: {
      type: String,
    },
    /**
     * Language for syntax highlighting
     */
    language: {
      type: String,
    },
  },
  data: function() {
    return {
      expanded: null,
      maxHeight: typeof window === 'undefined' ? undefined : null,
      copied: null,
      height: null,
      isExpandable: null,
    };
  },
  computed: {
    source() {
      if (this.base64) return Base64.decode(this.base64);
      return this.value;
    },
    out() {
      return this.$slots.default;
    },
  },
  mounted() {
    if (this.$refs.body) {
      this.isExpandable = this.$refs.body.scrollHeight > 1000;
      this.height = this.$refs.body.scrollHeight - 700;
      this.expanded = this.$refs.body.scrollHeight - 700 < 300;
      this.maxHeight = this.$refs.body.scrollHeight + "px";
    }
  },
  methods: {
    filename(url) {
      const tokens = url
        .replace(/\#.*$/, "")
        .split("/")
        .slice(7);
      if (tokens.length > 4) {
        return [
          tokens[0],
          tokens[1],
          "...",
          tokens.slice(-2)[0],
          tokens.slice(-2)[1],
        ].join(" / ");
      } else {
        return tokens.join(" / ");
      }
    },
    copy(value) {
      const val = value
        .replace(/&quot;/g, '"')
        .replace(/&lt;/g, "<")
        .replace(/&gt;/g, ">")
        .replace(/&amp;/g, "&");
      this.copied = true;
      copy(val);
      setTimeout(() => {
        this.copied = false;
      }, 2000);
    },
    highlighted(source) {
      const supportedSyntax = Prism.languages[this.language];
      if (supportedSyntax) {
        return Prism.highlight(
          source
            .replace(/&quot;/g, '"')
            .replace(/&lt;/g, "<")
            .replace(/&gt;/g, ">")
            .replace(/&amp;/g, "&"),
          supportedSyntax
        );
      } else {
        return source;
      }
    },
    expand(bool, scroll) {
      const container = this.$refs.container;
      this.expanded = bool;
      if (!bool && container && scroll) container.scrollIntoView();
    },
  },
};
</script>
