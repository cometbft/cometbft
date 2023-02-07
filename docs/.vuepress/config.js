module.exports = {
  theme: 'cosmos',
  title: 'CometBFT',
  base: process.env.VUEPRESS_BASE,
  themeConfig: {
    repo: 'cometbft/cometbft',
    docsRepo: 'cometbft/cometbft',
    docsDir: 'docs',
    editLinks: true,
    label: 'core',
    algolia: {
      id: "QQFROLBNZC",
      key: "f1b68b96fb31d8aa4a54412c44917a26",
      index: "cometbft"
    },
    versions: [
      {
        "label": "v0.34 (latest)",
        "key": "v0.34"
      },
      {
        "label": "v0.37",
        "key": "v0.37"
      }
    ],
    topbar: {
      banner: false,
    },
    sidebar: {
      auto: true,
      nav: [
        {
          title: 'Resources',
          children: [
            {
              title: 'RPC',
              path: (process.env.VUEPRESS_BASE ? process.env.VUEPRESS_BASE : '/')+'rpc/',
              static: true
            },
          ]
        }
      ]
    },
    gutter: {
      title: 'Help & Support',
      editLink: true,
      forum: {
        title: 'CometBFT Discussions',
        text: 'Join the CometBFT discussions to learn more',
        url: 'https://github.com/cometbft/cometbft/discussions',
        bg: '#0B7E0B',
        logo: 'cometbft'
      },
      github: {
        title: 'Found an Issue?',
        text: 'Help us improve this page by suggesting edits on GitHub.'
      }
    },
    footer: {
      question: {
        text: 'Chat with CometBFT developers in <a href=\'https://discord.gg/vcExX9T\' target=\'_blank\'>Discord</a> or reach out on <a href=\'https://github.com/cometbft/cometbft/discussions\' target=\'_blank\'>GitHub</a> to learn more.'
      },
      logo: '/logo-bw.svg',
      textLink: {
        text: 'cometbft.com',
        url: 'https://cometbft.com'
      },
      services: [
        {
          service: 'medium',
          url: 'https://medium.com/@cometbft'
        },
        {
          service: 'twitter',
          url: 'https://twitter.com/cometbft'
        },
        {
          service: 'linkedin',
          url: 'https://www.linkedin.com/company/informal-systems/'
        },
        {
          service: 'reddit',
          url: 'https://reddit.com/r/cosmosnetwork'
        },
        {
          service: 'telegram',
          url: 'https://t.me/cosmosproject'
        },
        {
          service: 'youtube',
          url: 'https://www.youtube.com/c/CosmosProject'
        }
      ],
      smallprint:
        'The development of CometBFT is led primarily by [Informal Systems](https://informal.systems/). Funding for this development comes primarily from the Interchain Foundation, a Swiss non-profit. The CometBFT trademark is owned by The Interchain Foundation.',
      links: [
        {
          title: 'Documentation',
          children: [
            {
              title: 'Cosmos SDK',
              url: 'https://docs.cosmos.network'
            },
            {
              title: 'Cosmos Hub',
              url: 'https://hub.cosmos.network'
            }
          ]
        },
        {
          title: 'Community',
          children: [
            {
              title: 'CometBFT blog',
              url: 'https://medium.com/@cometbft'
            },
            {
              title: 'GitHub Discussions',
              url: 'https://github.com/cometbft/cometbft/discussions'
            }
          ]
        },
        {
          title: 'Contributing',
          children: [
            {
              title: 'Contributing to the docs',
              url: 'https://github.com/cometbft/cometbft'
            },
            {
              title: 'Source code on GitHub',
              url: 'https://github.com/cometbft/cometbft'
            },
            {
              title: 'Careers at CometBFT',
              url: 'https://informal.systems/careers'
            }
          ]
        }
      ]
    }
  },
  plugins: [
    [
      '@vuepress/google-analytics',
      {
        ga: 'UA-51029217-11'
      }
    ],
    [
      '@vuepress/plugin-html-redirect',
      {
        countdown: 0
      }
    ]
  ]
};
