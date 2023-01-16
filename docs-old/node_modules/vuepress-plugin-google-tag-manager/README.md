# vuepress-plugin-google-tag-manager

> vuepress-plugin-google-tag-manager plugin for vuepress

## Install

```
npm i vuepress-plugin-google-tag-manager --save
```
or
```
yarn add vuepress-plugin-google-tag-manager
```

## Configuration

```javascript
module.exports = {
  plugins: ['vuepress-plugin-google-tag-manager'] 
}
```

## Options

### gtm

- Type: `string`
- Default: `undefined`

Provide the Google Tag Manager ID to enable integration.

## Documentation

Once the configuration is completed, you can access vue gtm instance in your components like that :

```javascript
export default {
  name: 'MyComponent',
  data() {
    return {
      someData: false
    };
  },
  methods: {
    onClick: function() {
      this.$gtm.trackEvent({
        event: null, // Event type [default = 'interaction'] (Optional)
        category: 'Calculator',
        action: 'click',
        label: 'Home page SIP calculator',
        value: 5000,
        noninteraction: false // Optional
      });
    }
  },
  mounted() {
    this.$gtm.trackView('MyScreenName', 'currentpath');
  }
};
```

The passed variables are mapped with GTM data layer as follows

```
dataLayer.push({
	'event': event || 'interaction',
	'target': category,
	'action': action,
	'target-properties': label,
	'value': value,
	'interaction-type': noninteraction,
	...rest
});
```