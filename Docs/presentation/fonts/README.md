# Presentation fonts

The presentation bundles the unmodified regular and bold font files used
for PDF builds:

- **Fira Sans:** [Google Fonts source](https://github.com/google/fonts/tree/main/ofl/firasans),
  with the original [OFL license](fira-sans/OFL.txt).
- **Fira Code:** files from the local `fonts-firacode` package,
  [upstream project](https://github.com/tonsky/FiraCode), with the package's
  [copyright notices and OFL license](fira-code/LICENSE.txt).

From the repository root, pass `--font-path Docs/presentation/fonts` to
Typst as shown in the [build instructions](../README.md#rebuild-the-pdfs).
The font binaries are included so builds do not depend on system-installed
copies or an additional font download.
