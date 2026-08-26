UPDATE settings
SET value = '',
    updated_at = NOW()
WHERE "group" = 'website_name'
  AND key IN ('body', 'website_name_body')
  AND (
      (locale = 'en' AND value = 'Content placeholder: this section will explain where the name comes from, what relationship it describes, and why it is different from “About Us”.')
      OR (locale = 'zh_cn' AND value = '内容占位：这里将写明这个名字的来源、它想表达的关系，以及它为什么不等同于“关于我们”。')
  );

UPDATE settings
SET value = '',
    updated_at = NOW()
WHERE "group" = 'website_name'
  AND key IN ('note', 'website_name_note')
  AND (
      (locale = 'en' AND value = 'The full text is being prepared')
      OR (locale = 'zh_cn' AND value = '正文内容正在整理中')
  );
