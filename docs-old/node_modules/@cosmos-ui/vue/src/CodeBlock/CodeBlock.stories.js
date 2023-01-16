import CodeBlock from "./CodeBlock.vue";
import { default as data } from "./data.js"

export default {
  title: "CodeBlock",
  component: CodeBlock,
};

export const normal = () => ({
  components: { CodeBlock },
  data: function () {
    return {
      data,
      url: "https://github.com/cosmos/sdk-tutorials/blob/c6754a1e313eb1ed973c5c91dcc606f2fd288811/deeply/nested/hidden/directory/go.mod#L1-L18",
      base64: "Ly8gVXBkYXRlIHRoZSB2YWxpZGF0b3Igc2V0CmZ1bmMgKGFwcCAqUGVyc2lzdGVudEtWU3RvcmVBcHBsaWNhdGlvbikgRW5kQmxvY2socmVxIHR5cGVzLlJlcXVlc3RFbmRCbG9jaykgdHlwZXMuUmVzcG9uc2VFbmRCbG9jayB7CglyZXR1cm4gdHlwZXMuUmVzcG9uc2VFbmRCbG9ja3tWYWxpZGF0b3JVcGRhdGVzOiBhcHAuVmFsVXBkYXRlc30KfQo="
    }
  },
  template: `
    <div>
      <p>One-line snippet without syntax highlighting:</p>
      <code-block :value="data.short"/>
      <p>Multiline snippet with syntax highlighting:</p>
      <code-block :value="data.medium" language="go"/>
      <p>Multiline snippet with an expand button:</p>
      <code-block :value="data.long" :url="url" language="xyz"/>
      <p>Base64:</p>
      <code-block :base64="base64" language="go"/>
      <p>Rust source:</p>
      <code-block :value="data.rust" language="rust"/>
      <p>Solidity source:</p>
      <code-block :value="data.solidity" language="solidity"/>
      <p>Python prismjs package is not imported:</p>
      <code-block :value="data.python" language="python"/>
    </div>
  `
});
