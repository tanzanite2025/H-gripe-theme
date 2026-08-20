DO $$
DECLARE
    restored_flow_id BIGINT;
BEGIN
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
    ON CONFLICT (slug) DO UPDATE SET
        name = EXCLUDED.name,
        description = EXCLUDED.description,
        product_category_slug = EXCLUDED.product_category_slug,
        is_enabled = TRUE,
        sort_order = EXCLUDED.sort_order,
        updated_at = NOW()
    RETURNING id INTO restored_flow_id;

    UPDATE selection_assistant_flow_versions
    SET
        status = 'archived',
        updated_at = NOW()
    WHERE flow_id = restored_flow_id
      AND status = 'published';

    INSERT INTO selection_assistant_flow_versions (
        flow_id,
        version_number,
        status,
        config,
        published_at
    )
    VALUES (
        restored_flow_id,
        1,
        'published',
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
              ]
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
              ]
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
              ]
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
              ]
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
              ]
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
              ]
            },
            {
              "key": "support_review",
              "type": "support",
              "prompt": {
                "en": "We have a starting fit profile. Continue with the next fit question or contact support.",
                "zh_cn": "我们已经得到一份初步适配信息，可以继续补充规格，也可以直接联系客服。"
              }
            }
          ]
        }'::jsonb,
        NOW()
    )
    ON CONFLICT (flow_id, version_number) DO UPDATE SET
        status = 'published',
        config = EXCLUDED.config,
        published_at = COALESCE(selection_assistant_flow_versions.published_at, EXCLUDED.published_at),
        updated_at = NOW();
END $$;
