# ADF-Converter E2E Test — Jira

Dieses Dokument dient als E2E-Testvorlage. Jede Section ist ein eigener
Test-Case. Mit jira-edit in ein Test-Ticket pasten und in Jira prüfen.

---

## 1. Paragraph + Basic Marks

Normaler Text ohne Formatierung.

**Bold text** und *italic text* und `inline code` hier.

~~Durchgestrichen~~ und <u>unterstrichen</u> kombiniert.

**Bold mit *italic nested* drin** — verschachtelte Marks.

## 2. Headings

# Heading 1
## Heading 2
### Heading 3
#### Heading 4
##### Heading 5
###### Heading 6

## 3. Links

[Einfacher Link](https://example.com)

[Link mit Title](https://example.com "Tooltip-Text")

Text mit [inline Link](https://atlassian.com) mitten im Satz.

## 4. Code Block

```javascript
function greet(name) {
  console.log(`Hello, ${name}!`);
  return true;
}
```

```python
def fibonacci(n):
    if n <= 1:
        return n
    return fibonacci(n-1) + fibonacci(n-2)
```

```
Code ohne Sprachangabe
```

## 5. Horizontal Rule

Text vor der Linie.

---

Text nach der Linie.

## 6. Blockquote

> Einfaches Zitat in einem Absatz.

> Verschachteltes Zitat:
>
> > Inneres Zitat

## 7. Bullet List

- Erster Punkt
- Zweiter Punkt
  - Verschachtelter Punkt
  - Noch einer
- Dritter Punkt

## 8. Ordered List

1. Erster Schritt
2. Zweiter Schritt
3. Dritter Schritt
   1. Sub-Schritt A
   2. Sub-Schritt B

## 9. Ordered List mit Start-Nummer

3. Startet bei drei
4. Vierter Eintrag
5. Fuenfter Eintrag

## 10. Task List

- [ ] Offene Aufgabe
- [x] Erledigte Aufgabe
- [ ] Noch eine offene

## 11. Table

| Header 1 | Header 2 | Header 3 |
|----------|----------|----------|
| Zelle 1  | Zelle 2  | Zelle 3  |
| **Bold** | *Italic* | `Code`   |
| Leer     |          | Wert     |

## 12. Panel (Fenced Div)

:::info
Das ist ein Info-Panel mit **formatiertem** Text.
:::

:::warning
Warnung: Hier stimmt etwas nicht!
:::

:::error
Fehler: Kritisches Problem aufgetreten.
:::

:::success
Erfolgreich abgeschlossen.
:::

:::note
Eine Notiz zur Kenntnisnahme.
:::

## 13. Expand (Details)

<details data-adf-type="expand">
  <summary>Klick mich auf</summary>
  Versteckter Inhalt im Expand-Block.
</details>

<details open data-adf-type="expand">
  <summary>Bereits offen</summary>
  Dieser Expand ist standardmaessig offen.
</details>

## 14. Mention

[@Max Mustermann](accountid:5b10ac8d82e05b22cc7d4ef5)

Text mit [@Anna Schmidt](accountid:abc123?accessLevel=CONTAINER&userType=DEFAULT) inline.

## 15. Emoji

Standard-Emojis: 😀 👍 ❤️ 🎉 ✅

## 16. Date

[date:2026-04-04]

Deadline ist [date:2026-12-31] fuer dieses Projekt.

## 17. Status

[status:In Progress|blue]

[status:Done|green]

[status:Blocked|red]

[status:On Hold|yellow]

[status:Review|purple]

[status:Draft|neutral]

## 18. Inline Card

[https://github.com/example/repo](https://github.com/example/repo)

## 19. Text Color

<span style="color: #ff0000">Roter Text</span> und <span style="color: #00ff00">gruener Text</span> und <span style="color: #0000ff">blauer Text</span>.

## 20. Subscript und Superscript

H<sub>2</sub>O ist Wasser.

E = mc<sup>2</sup> ist Einsteins Formel.

## 21. Kombinierte Marks

**<u>Bold und unterstrichen</u>** zusammen.

*~~Italic und durchgestrichen~~* kombiniert.

<span style="color: #ff5500">**Farbiger Bold-Text**</span> verschachtelt.

## 22. Komplexer Absatz

Hier kommt alles zusammen: **Bold**, *italic*, `code`, ~~strike~~, <u>underline</u>,
[ein Link](https://example.com), <span style="color: #ff0000">farbig</span>,
H<sub>2</sub>O und x<sup>2</sup> — alles in einem Absatz.
