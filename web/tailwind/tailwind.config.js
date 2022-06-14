module.exports = {
    content: [
        "../html/**/*.{vue,js,ts,jsx,tsx,html}",
    ],
    theme: {
        extend: {},
    },
    plugins: [
        require('@tailwindcss/forms')
    ]
}