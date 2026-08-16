# Wheelset Fit Questionnaire Specification

Last updated: 2026-08-16

## 1. Decision

The wheelset fit experience is one dedicated, fixed-order questionnaire for customers who do not know which wheelset specifications fit their bike.

It is not QUICKBUY/DIY, a general-purpose workflow engine, a decision graph, or a collection of separate assistants.

The first production questionnaire is `wheelset-fit`. It always starts from the system product category slug `wheelset`, which may move within the product-category tree but must not be deleted, disabled, or renamed.

The questionnaire contract is:

```text
ordered questions
  -> selected answers
  -> wheelset product filters
  -> recommended products
  -> structured customer-service and AI request
```

Questions are shown one at a time in the storefront modal. The product-result panel updates after every answer. The question order is fixed by `sort_order`; there are no graph nodes, edges, coordinates, or `next_node_key` values.

## 2. Product Boundary

### Wheelset Fit Questionnaire

This system answers:

- What wheelset specifications does this customer know?
- Which compatible wheelset products remain after each confirmed answer?
- Which details remain unknown and need staff or AI follow-up?
- What structured fit profile should be sent to customer service?

It serves customers who need help choosing a compatible complete wheelset.

### QUICKBUY / DIY

QUICKBUY is a separate configurable-purchase workflow for customers who already know which components they want to buy or assemble.

Its contract remains:

```text
purchase steps
  -> product candidates
  -> selected products
  -> cart or order assembly
```

The questionnaire may be opened from a QUICKBUY entry point, but the integration is output-only. QUICKBUY must not own questionnaire questions, answers, translations, filters, draft state, or published versions.

Allowed handoff:

```text
fit answers
product query
unknown details
structured support request
selected product
```

## 3. Storefront Interaction

### 3.1 Fixed question progression

The customer sees one current question at a time. Selecting an answer advances to the next enabled question in `sort_order`. Back returns to the preceding answered question and removes subsequent answers before recalculating the product query.

The questionnaire must not ask the customer to type a bicycle model, frame model, free-form specification, or technical measurement as the primary path. Use clear, finite choices. Every question that can reasonably be unknown must include an explicit `unknown` option.

Example first questions:

```text
1. Wheelset type
   - Mountain 26
   - Mountain 27.5
   - Mountain 29
   - Road disc 700C
   - Road V-brake 700C
   - Snow 26
   - Small wheel 20
   - I am not sure

2. Front axle standard
3. Rear axle standard
4. Brake rotor mount
5. Freehub body
```

`Mountain 29` is one answer option with a stable technical value. It is not a branch to another assistant, and it does not create a second questionnaire.

### 3.2 HELP is part of every question

Each question may include a customer-facing HELP explanation. HELP explains why the information matters, how to identify it, and what to do when the customer does not know.

The question surface exposes HELP through a standard question-mark button. On mobile it opens a focused sheet or dialog; on desktop it may open a dialog or anchored panel. HELP must not consume the option area by default.

Required HELP fields when HELP is present:

```text
help_title
help_body
```

Example:

```text
Question: Rear axle standard

HELP title: Why do we need this?
HELP body: The rear axle standard must match the rear dropouts on your frame.
Check the frame manual or the marking on the existing axle. If you are not sure,
choose "I am not sure" and we will keep it for customer service or AI follow-up.
```

HELP is educational copy. It does not contain product-filter rules, hidden business logic, or unresolved technical assumptions.

### 3.3 Product results and support handoff

The result query always begins with:

```json
{
  "category_slug": "wheelset"
}
```

Each selected answer can add product specification filters. Unknown answers add no restrictive filter for that fact and are recorded in `unknowns`.

Customers may open customer service at any point. The support request must include the currently answered stable values, labels in the visitor locale, unanswered required facts, the questionnaire version, and the derived wheelset product query.

Example stable output:

```json
{
  "questionnaire_slug": "wheelset-fit",
  "questionnaire_version": 3,
  "answers": {
    "wheel_standard": "mtb_29",
    "front_axle": "15x110",
    "rear_axle": "12x148",
    "brake_mount": "center_lock",
    "freehub": "micro_spline"
  },
  "unknowns": [],
  "product_query": {
    "category_slug": "wheelset",
    "spec_filters": {
      "wheel_size": ["29"],
      "front_axle": ["15x110"],
      "rear_axle": ["12x148"],
      "brake_mount": ["center_lock"],
      "freehub": ["micro_spline"]
    }
  }
}
```

The display labels are never the technical identity. Staff, AI, analytics, product filtering, and historical records must rely on stable answer values and keys.

## 4. Domain Model

The questionnaire remains versioned so storefront customers only read a validated published version, while Admin edits a draft. The question model is ordered, not graph-shaped.

```text
wheelset_fit_questionnaires
  -> wheelset_fit_questionnaire_versions
  -> wheelset_fit_questions
  -> wheelset_fit_question_options
  -> question and option translations
```

### 4.1 Questionnaire and version

`wheelset_fit_questionnaires` holds the single logical questionnaire identity.

```text
id
slug                         -- fixed: wheelset-fit
product_category_slug        -- fixed: wheelset
is_enabled
created_at
updated_at
```

`wheelset_fit_questionnaire_versions` holds immutable published content and one editable draft.

```text
id
questionnaire_id
version_number
status                       -- draft / published / archived
published_at
published_by
created_at
updated_at
```

Rules:

- At most one published version exists for the questionnaire.
- Publishing archives the previous published version.
- A published version is never edited in place.
- Every support request, analytics event, and future AI output records the published version used.
- A storefront request must never receive a draft.

### 4.2 Questions

`wheelset_fit_questions` defines a question once per questionnaire version.

```text
id
questionnaire_version_id
question_key                 -- stable machine identity, for example rear_axle
answer_key                   -- stable output field, normally equal to question_key
sort_order
input_mode                   -- single_choice initially
is_required
allow_unknown
is_enabled
created_at
updated_at
```

Rules:

- `question_key` is lowercase snake_case and must be unique within a questionnaire version.
- `sort_order` is the only customer progression rule.
- `input_mode` starts with `single_choice`. Do not add free text, graph jumps, or arbitrary custom input without a separate approved specification.
- A required question with `allow_unknown = true` is completed when its explicit unknown option is chosen.
- Disable or delete a question only in a draft. Historical published versions remain intact.

### 4.3 Answer options

`wheelset_fit_question_options` defines the selectable answers for a question.

```text
id
question_id
option_key                   -- stable local identity, for example mtb_29
answer_value                 -- stable technical value, for example mtb_29
sort_order
is_unknown
is_enabled
product_filter_effects_json
created_at
updated_at
```

`product_filter_effects_json` is a structured mapping from an answer to existing product specification filters.

```json
{
  "spec_filters": {
    "wheel_size": ["29"],
    "riding_category": ["mtb"]
  }
}
```

Rules:

- `option_key` and `answer_value` are technical identifiers, not translated labels.
- An option may set more than one product specification filter.
- An unknown option must set `is_unknown = true` and must not add restrictive product filters for the unanswered fact.
- Product-filter keys and values must be validated against the product specification contract before a version can be published.
- Options do not point to another question. There is no `next_node_key` field.

## 5. Translation Model

The questionnaire uses the existing storefront locale registry as the only supported language source. Current supported locale codes are owned by `shared/storefront-locales.json` and must not be duplicated in questionnaire code or Admin UI.

Translations are first-class records, not maps embedded inside a JSON graph.

```text
wheelset_fit_question_translations
  question_id
  locale
  prompt
  help_title
  help_body
  source_revision
  translated_revision
  updated_at

wheelset_fit_question_option_translations
  option_id
  locale
  label
  description
  source_revision
  translated_revision
  updated_at
```

The unique identity for every translation row is `(parent_id, locale)`.

### 5.1 Translation completion states

Translation status is derived, not manually declared:

- `complete`: required text exists and `translated_revision` equals the source revision.
- `missing`: no required translation text exists.
- `outdated`: source content changed after this translation was last confirmed.

The initial source locale is `zh_cn` for this questionnaire. The source locale is a workflow choice, not a fallback rule.

When a source question, HELP text, option label, or option description changes:

1. Increase that content item's source revision.
2. Mark every other locale translation for that item as `outdated`.
3. Keep translation records and option identities in place.
4. Require explicit human confirmation before the translation becomes `complete` again.

Structural synchronization is automatic: adding an option adds a translation slot for every supported locale. Translation text is not automatically copied as a completed translation. A future machine-translation action may create a draft, but it must not silently mark a locale complete.

### 5.2 Draft save data flow

The service saves one complete question aggregate into the current draft. The input always includes the structural question fields. Translation and option behavior follows these patch rules:

1. `GetOrCreateDraft` returns an existing draft unchanged. When no draft exists, it creates the next version and deep-copies every question, option, and translation from the latest published version. If nothing has been published, the first draft is empty.
2. `SaveQuestion` treats submitted translations as locale patches. It retains every omitted locale's current text, then materializes a row for every enabled storefront locale before persistence.
3. A source-locale text change increments only that question or option's `source_revision`. It updates every locale row's `source_revision`; omitted non-source translations remain in place and become outdated because their `translated_revision` no longer matches.
4. A submitted non-source translation becomes complete only when its required fields are present; then its `translated_revision` is set to the current source revision. An incomplete submission remains missing.
5. `options` omitted from an existing question means retain the complete current option set. An explicit empty array means replace the option set with no options. When options are supplied, the set is replaced and existing option identities are matched by `option_key`.
6. Persistence locks the draft version, reloads the aggregate, performs the merge, and upserts translations by `(parent_id, locale)` in one transaction. Saves and publishing of the same draft are therefore serialized; a later save merges from the previous committed state instead of overwriting it from an earlier snapshot.

The repository owns transactions, draft-state protection, version cloning, and relational upserts. The service owns input normalization, locale completion, merge behavior, and revision calculation. A future Admin handler must call the service rather than writing the repository directly.

### 5.3 Runtime locale policy

- The public API receives the visitor's canonical storefront locale.
- A published questionnaire must return text only for that locale.
- Missing questionnaire translations must not silently render unrelated English or Chinese customer-facing copy.
- Before enabling the questionnaire for a locale, the published version must pass completeness validation for that locale.
- Administrative previews may show a clear missing-translation state; public storefront traffic may not.

This prevents a visitor from seeing an answer label in one language and its HELP explanation in another.

## 6. Admin Experience

There is one Admin tab: `轮组选型问卷`. It manages one questionnaire and does not display an assistant list or an add-assistant action.

### 6.1 Main page

The main page contains:

- Questionnaire status: enabled state, published version, pending draft, question count.
- Draft and published lifecycle actions: save draft, validate, publish.
- One ordered question list.
- A single question editor dialog.

The question list shows compact operational facts only:

```text
01  Wheelset type          8 answers   HELP ready   20/20 locales complete
02  Front axle standard    5 answers   HELP ready   16/20 locales complete
03  Rear axle standard     5 answers   HELP missing  0/20 locales complete
```

It must not show a canvas, graph nodes, graph lines, node coordinates, targets, or branch navigation controls.

### 6.2 Question editor dialog

Selecting a question opens a full dialog with three clear sections or tabs.

```text
Basic
  question key
  answer key
  order
  required state
  allow unknown state
  enabled state

Answers and product matching
  answer option list
  option key
  stable answer value
  unknown state
  product specification filter effects

Translations
  controlled storefront locale selector
  prompt
  HELP title
  HELP body
  option labels and descriptions
  completion status for the selected locale
```

The translation section must reuse the existing Admin storefront-locale selector and supported-language composable. It must not add a page-local locale list or free-text locale field.

The dialog keeps one selected locale at a time. It does not render 20 textareas for every question. A locale summary shows complete, missing, and outdated counts, and selecting a locale loads the same question structure with that locale's text fields.

### 6.3 Authoring constraints

- Product filter effects are selected from existing product specification keys and allowed values where possible. Free-form filter-key fields are prohibited.
- An option cannot be deleted when doing so would invalidate a published version; create a new draft first.
- Reordering questions changes only future published versions.
- HELP is optional at the database level but required by product policy for any question that asks a customer to identify a frame, axle, brake, drivetrain, or other technical compatibility fact.
- A question editor preview renders the visitor-facing question, question-mark HELP affordance, and answer buttons for the selected locale.

## 7. Validation and Publication

Draft validation must reject:

- More than one questionnaire identity or a product category other than `wheelset`.
- Duplicate or invalid question keys.
- Duplicate or invalid option keys within a question.
- Missing or duplicate sort order.
- A question with no enabled options.
- A required question with no valid unknown option when `allow_unknown = true`.
- An unknown option that adds a restrictive filter for its unanswered answer key.
- Invalid product specification filter keys or values.
- Missing required source-locale text.
- Missing or outdated translations for any locale selected for public enablement.
- A technical question without required HELP text.

Validation does not need graph checks such as cycles, unreachable nodes, or missing next-node targets because the domain model has no graph edges.

Publishing performs validation against the exact draft snapshot, archives the previous published version atomically, invalidates the public questionnaire cache, and records the publisher identity and timestamp.

## 8. Storefront API Contract

The public API returns one published questionnaire version in the requested locale.

```http
GET /api/v1/wheelset-fit/questionnaire?locale=zh_cn
```

Response shape:

```json
{
  "data": {
    "slug": "wheelset-fit",
    "version": 3,
    "category_slug": "wheelset",
    "questions": [
      {
        "key": "wheel_standard",
        "answer_key": "wheel_standard",
        "sort_order": 10,
        "required": true,
        "allow_unknown": true,
        "prompt": "请选择您的轮组基础类型",
        "help": {
          "title": "为什么需要这个？",
          "body": "轮径和制动形式决定可匹配的轮组范围。"
        },
        "options": [
          {
            "key": "mtb_29",
            "value": "mtb_29",
            "label": "山地 29",
            "description": "适用于 29 英寸山地车架"
          }
        ]
      }
    ]
  }
}
```

The public response does not expose Admin translation status, source revisions, unpublished content, arbitrary editor metadata, or raw product-filter rules unless the storefront must use a normalized, validated query effect supplied by the backend.

The storefront derives the product query from selected stable answers through a dedicated mapping boundary. The browser must not invent technical compatibility rules from display text.

## 9. Customer-Service and AI Contract

The customer-service message type remains a structured wheelset-fit request. Its payload must use stable answer keys and values plus localized display labels.

```json
{
  "kind": "wheelset_fit_request",
  "questionnaire_slug": "wheelset-fit",
  "questionnaire_version": 3,
  "source": "guides/wheelset-buyers",
  "answers": {
    "wheel_standard": "mtb_29",
    "front_axle": "unknown",
    "rear_axle": "12x148"
  },
  "answer_labels": {
    "wheel_standard": "山地 29",
    "front_axle": "不确定",
    "rear_axle": "12x148 Boost"
  },
  "unknowns": ["front_axle"],
  "product_query": {
    "category_slug": "wheelset",
    "spec_filters": {
      "wheel_size": ["29"],
      "rear_axle": ["12x148"]
    }
  }
}
```

Customer-service UI and future AI consumers read this payload as a fact sheet. They must not need to interpret question order, UI routes, or a graph traversal path.

## 10. Migration From The Retired Graph Model

The current graph-based selection assistant is a temporary implementation and must not receive new questions or new editor features.

Retired concepts:

```text
selection assistant list
assistant creation
graph nodes
graph editor coordinates
node type
next_node_key
cycle validation
unreachable-node validation
branch traversal
```

Migration sequence:

1. Create the questionnaire tables and locale-controlled translation tables.
2. Seed one draft `wheelset-fit` questionnaire from the current published content, preserving technical answer values and product filter effects.
3. Build the Admin ordered question list and question editor dialog before exposing the new storefront runtime.
4. Validate the source locale and each enabled storefront locale.
5. Publish the questionnaire version.
6. Switch the storefront modal and customer-service payload builder to the new public questionnaire endpoint.
7. Retire the graph Admin tab, graph API, graph runtime types, and graph tables only after the new published questionnaire has been verified.

Do not make the two models share an editor surface, version row, or runtime contract. The questionnaire is a new bounded context because it intentionally has a narrower domain model.

## 11. Non-Goals

- Do not support arbitrary conditional graph navigation.
- Do not create one assistant per riding scenario.
- Do not make customers enter bicycle models as the normal path.
- Do not use free-text input as a substitute for missing product data.
- Do not place all 20 locale text fields on one Admin screen.
- Do not treat copied source text as a completed translation.
- Do not make QUICKBUY own questionnaire behavior.
- Do not use product titles, translated labels, or HELP text as technical filter identities.

## 12. Implementation Update Rule

Any change to question ordering, answer semantics, HELP behavior, supported locales, translation completion, product-filter mapping, public API, support payload, or version publication must update this document in the same change.
