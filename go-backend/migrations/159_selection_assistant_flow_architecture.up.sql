CREATE TABLE IF NOT EXISTS selection_assistant_flows (
    id BIGSERIAL PRIMARY KEY,
    slug VARCHAR(120) NOT NULL UNIQUE,
    name VARCHAR(160) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    product_category_slug VARCHAR(120) NOT NULL DEFAULT 'wheelset',
    is_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    sort_order INTEGER NOT NULL DEFAULT 100,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT ck_selection_assistant_flows_slug
        CHECK (slug ~ '^[a-z0-9]+([_-][a-z0-9]+)*$'),
    CONSTRAINT ck_selection_assistant_flows_product_category
        CHECK (product_category_slug = 'wheelset')
);

CREATE INDEX IF NOT EXISTS idx_selection_assistant_flows_enabled_order
    ON selection_assistant_flows(is_enabled, sort_order, id);

CREATE TABLE IF NOT EXISTS selection_assistant_flow_versions (
    id BIGSERIAL PRIMARY KEY,
    flow_id BIGINT NOT NULL REFERENCES selection_assistant_flows(id) ON DELETE CASCADE,
    version_number INTEGER NOT NULL DEFAULT 1,
    status VARCHAR(24) NOT NULL DEFAULT 'draft',
    config JSONB NOT NULL DEFAULT '{}'::jsonb,
    published_at TIMESTAMPTZ,
    published_by BIGINT REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT ck_selection_assistant_flow_versions_status
        CHECK (status IN ('draft', 'published', 'archived')),
    CONSTRAINT ck_selection_assistant_flow_versions_number
        CHECK (version_number > 0)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_selection_assistant_versions_flow_number
    ON selection_assistant_flow_versions(flow_id, version_number);

CREATE UNIQUE INDEX IF NOT EXISTS idx_selection_assistant_versions_one_published
    ON selection_assistant_flow_versions(flow_id)
    WHERE status = 'published';

CREATE INDEX IF NOT EXISTS idx_selection_assistant_versions_status
    ON selection_assistant_flow_versions(flow_id, status, version_number DESC);

INSERT INTO selection_assistant_flows (
    slug,
    name,
    description,
    product_category_slug,
    is_enabled,
    sort_order
)
VALUES (
    'wheelset-fit-helper',
    '轮组适配助手',
    '通过分支问答收集轮组适配信息，并将结构化结果交给商品搜索和客服对话。',
    'wheelset',
    TRUE,
    100
)
ON CONFLICT (slug) DO NOTHING;

INSERT INTO selection_assistant_flow_versions (
    flow_id,
    version_number,
    status,
    config
)
SELECT
    flow.id,
    1,
    'draft',
    '{
      "kind": "product_selection_assistant",
      "schema_version": 1,
      "entry_node_key": "riding_category",
      "base_product_query": {
        "category_slug": "wheelset"
      },
      "nodes": [
        {
          "key": "riding_category",
          "type": "question",
          "prompt": {
            "en": "Choose your main riding category",
            "zh_cn": "请选择您的主要骑行类别"
          },
          "helper": {
            "en": "We will only show the next question for the branch you choose.",
            "zh_cn": "选择后只显示该分支的下一项问题。"
          },
          "options": [
            {
              "key": "mtb",
              "label": {
                "en": "Mountain",
                "zh_cn": "山地"
              },
              "answer_effects": {
                "riding_category": "mtb"
              },
              "next_node_key": "mtb_wheel_size"
            },
            {
              "key": "road_disc",
              "label": {
                "en": "Road disc brake",
                "zh_cn": "公路碟刹"
              },
              "answer_effects": {
                "riding_category": "road_disc"
              },
              "next_node_key": "road_disc_wheel_size"
            },
            {
              "key": "snow",
              "label": {
                "en": "Snow",
                "zh_cn": "雪地"
              },
              "answer_effects": {
                "riding_category": "snow"
              },
              "next_node_key": "snow_wheel_size"
            },
            {
              "key": "road_v_brake",
              "label": {
                "en": "Road V-brake",
                "zh_cn": "V 刹公路"
              },
              "answer_effects": {
                "riding_category": "road_v_brake"
              },
              "next_node_key": "road_v_brake_wheel_size"
            },
            {
              "key": "small_wheel",
              "label": {
                "en": "Small wheel",
                "zh_cn": "小轮"
              },
              "answer_effects": {
                "riding_category": "small_wheel"
              },
              "next_node_key": "small_wheel_size"
            }
          ],
          "editor": {
            "x": 80,
            "y": 140
          }
        },
        {
          "key": "mtb_wheel_size",
          "type": "question",
          "prompt": {
            "en": "What wheel size does your MTB frame support?",
            "zh_cn": "您的山地车架支持哪种轮径？"
          },
          "options": [
            {
              "key": "26",
              "label": {
                "en": "26",
                "zh_cn": "26"
              },
              "answer_effects": {
                "wheel_size": "26"
              },
              "query_effects": {
                "spec_filters": {
                  "wheel_size": ["26"]
                }
              },
              "next_node_key": "support_review"
            },
            {
              "key": "27_5",
              "label": {
                "en": "27.5",
                "zh_cn": "27.5"
              },
              "answer_effects": {
                "wheel_size": "27.5"
              },
              "query_effects": {
                "spec_filters": {
                  "wheel_size": ["27.5"]
                }
              },
              "next_node_key": "support_review"
            },
            {
              "key": "29",
              "label": {
                "en": "29",
                "zh_cn": "29"
              },
              "answer_effects": {
                "wheel_size": "29"
              },
              "query_effects": {
                "spec_filters": {
                  "wheel_size": ["29"]
                }
              },
              "next_node_key": "support_review"
            }
          ],
          "editor": {
            "x": 430,
            "y": 40
          }
        },
        {
          "key": "road_disc_wheel_size",
          "type": "question",
          "prompt": {
            "en": "What wheel size does your road frame support?",
            "zh_cn": "您的公路车架支持哪种轮径？"
          },
          "options": [
            {
              "key": "700c",
              "label": {
                "en": "700C",
                "zh_cn": "700C"
              },
              "answer_effects": {
                "wheel_size": "700c"
              },
              "query_effects": {
                "spec_filters": {
                  "wheel_size": ["700c"]
                }
              },
              "next_node_key": "support_review"
            }
          ],
          "editor": {
            "x": 430,
            "y": 240
          }
        },
        {
          "key": "snow_wheel_size",
          "type": "question",
          "prompt": {
            "en": "What wheel size does your snow bike support?",
            "zh_cn": "您的雪地车支持哪种轮径？"
          },
          "options": [
            {
              "key": "26",
              "label": {
                "en": "26",
                "zh_cn": "26"
              },
              "answer_effects": {
                "wheel_size": "26"
              },
              "query_effects": {
                "spec_filters": {
                  "wheel_size": ["26"]
                }
              },
              "next_node_key": "support_review"
            }
          ],
          "editor": {
            "x": 430,
            "y": 440
          }
        },
        {
          "key": "road_v_brake_wheel_size",
          "type": "question",
          "prompt": {
            "en": "What wheel size does your V-brake road frame support?",
            "zh_cn": "您的 V 刹公路车架支持哪种轮径？"
          },
          "options": [
            {
              "key": "700c",
              "label": {
                "en": "700C",
                "zh_cn": "700C"
              },
              "answer_effects": {
                "wheel_size": "700c"
              },
              "query_effects": {
                "spec_filters": {
                  "wheel_size": ["700c"]
                }
              },
              "next_node_key": "support_review"
            }
          ],
          "editor": {
            "x": 430,
            "y": 640
          }
        },
        {
          "key": "small_wheel_size",
          "type": "question",
          "prompt": {
            "en": "What wheel size does your small-wheel bike support?",
            "zh_cn": "您的小轮车支持哪种轮径？"
          },
          "options": [
            {
              "key": "20",
              "label": {
                "en": "20",
                "zh_cn": "20"
              },
              "answer_effects": {
                "wheel_size": "20"
              },
              "query_effects": {
                "spec_filters": {
                  "wheel_size": ["20"]
                }
              },
              "next_node_key": "support_review"
            }
          ],
          "editor": {
            "x": 430,
            "y": 840
          }
        },
        {
          "key": "support_review",
          "type": "support",
          "prompt": {
            "en": "We have a starting fit profile. Continue with the next fit question or contact support.",
            "zh_cn": "我们已经得到一份初步适配信息，可以继续补充规格，也可以直接联系客服。"
          },
          "editor": {
            "x": 790,
            "y": 400
          }
        }
      ]
    }'::jsonb
FROM selection_assistant_flows AS flow
WHERE flow.slug = 'wheelset-fit-helper'
  AND NOT EXISTS (
      SELECT 1
      FROM selection_assistant_flow_versions AS version
      WHERE version.flow_id = flow.id
  );
