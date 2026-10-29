# 46. The note of an invitation carries bold, italics and lists

- Status: Accepted
- Date: 2026-10-29

## Context

An invitation to read a Window goes by e-mail with a note of up to 500 characters in plain text, which the message shows apart from the words of Cahut, as the words of whoever wrote it ([ADR 31](0031-e-mail-through-a-relay.md)). The page of a personal link shows it too. In the console, the note is a bare text area.

People who invite want to set a few words in bold, or to list what to read first, as any mail client lets them. AIs that call `invite_to_window` write Markdown already, and their stars went out as they were.

The HTML of a message is Cahut's own screen, with its brand ([ADR 12](0012-website-and-docs-as-astro-sites.md)): nothing in it comes from the internet, and its words are plain text, which the layout escapes. The note is the one text in it that someone else wrote.

## Decision

- A note is plain text with a few marks of Markdown: `**bold**`, `*italics*`, `***both***`, and lists whose lines start with `-`, `*` or `+`, or with a number and a period. A blank line ends a paragraph or a list, a line break stays a line break, and a backslash before a mark keeps it as it is.
- A star opens before a word and closes after one, never next to a space, and a star that nothing closes stays a star: `5 * 3` reads as written. Nothing else is markup: no link, heading, image, code or HTML.
- The server reads a note into paragraphs and lists, once, in Go. The HTML of the e-mail sets them in `strong`, `em`, `ul` and `ol`, from text that the layout escapes: no markup of the note goes into the message. The plain text part quotes the note with the marks of its styles, without the backslashes.
- `get_invitation` gives the note as written, and in its paragraphs and lists, which the page of the invitation shows. The console does not read the marks itself.
- The console offers a small editor in place of the text area: a toolbar with bold, italics, a bulleted list and a numbered list, with the keyboard shortcuts of mail clients, over the text, which shows the styles as they will read. It writes the marks, and a backslash before a star or a start of line that would read as one. What is pasted comes in as plain text, and the editor refuses the styles that a note does not have, such as underline.
- A note holds 500 characters at most, its marks included, as the server counts them. The editor counts them as it goes, and the console does not send a longer note.
- `invite_to_window` tells AIs the marks in the description of `note`.

## Consequences

- People and AIs set the same few styles, which every mail client shows, and the plain text reads as the person wrote it.
- The notes of invitations sent before read the same, but for the stars that they held, which now set italics or bold.
- A marked note holds fewer words: the marks count in the 500 characters.
- Links stay as they were: a web address in a note is text, which mail clients may make clickable, as ADR 31 accepted. A link whose words hide its address, in an e-mail from the domain of a Cahut, is the shape of phishing, and the note does not offer it.
- The editor uses `contenteditable` and the editing commands of the browser, which are old but work in every browser that the console supports. It reads the styles that the browser shows rather than the markup it made, so that the marks follow what the person saw. No package is added.

## Alternatives considered

- **Markdown in the text area, with a preview.** Simpler and sure, but people would read stars while they write, and a preview doubles the field.
- **Restricted HTML for the note.** What a rich editor makes, but every client of the API, AIs included, would write HTML, and the server would have to clean it: a sanitizer is a dependency or a risk.
- **A rich text package such as ProseMirror or Tiptap.** A sturdier editor, for tens of kilobytes and a dependency, to set four styles in 500 characters.
- **Links with their own words.** Useful, but see above: the address of a link shows only as text.
