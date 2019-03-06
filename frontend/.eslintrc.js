module.exports = {
    root: true,
    env: {
        node: true
    },
    'extends': [
        'eslint:recommended',
        'plugin:vue/recommended',
        '@vue/standard'
    ],
    rules: {
        'no-console': process.env.NODE_ENV === 'production' ? 'warn' : 'off',
        'no-debugger': process.env.NODE_ENV === 'production' ? 'error' : 'off'
    },
    parserOptions: {
        parser: 'babel-eslint'
    }
};
