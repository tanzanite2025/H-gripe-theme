DO $$
DECLARE
    target_questionnaire_id BIGINT;
    seed_version_id BIGINT;
    next_version_number INTEGER;
    published_count INTEGER;
    question_count INTEGER;
    riding_question_id BIGINT;
    wheel_size_question_id BIGINT;
    option_id BIGINT;
BEGIN
    SELECT id
    INTO target_questionnaire_id
    FROM wheelset_fit_questionnaires
    WHERE slug = 'wheelset-fit';

    IF target_questionnaire_id IS NULL THEN
        INSERT INTO wheelset_fit_questionnaires (
            slug,
            product_category_slug,
            source_locale,
            is_enabled
        )
        VALUES (
            'wheelset-fit',
            'wheelset',
            'zh_cn',
            TRUE
        )
        RETURNING id INTO target_questionnaire_id;
    END IF;

    SELECT COUNT(*)
    INTO published_count
    FROM wheelset_fit_questionnaire_versions
    WHERE questionnaire_id = target_questionnaire_id
      AND status = 'published';

    IF published_count = 0 THEN
        SELECT id
        INTO seed_version_id
        FROM wheelset_fit_questionnaire_versions
        WHERE questionnaire_id = target_questionnaire_id
          AND status = 'draft'
        ORDER BY version_number DESC, id DESC
        LIMIT 1;

        IF seed_version_id IS NULL THEN
            SELECT COALESCE(MAX(version_number), 0) + 1
            INTO next_version_number
            FROM wheelset_fit_questionnaire_versions
            WHERE questionnaire_id = target_questionnaire_id;

            INSERT INTO wheelset_fit_questionnaire_versions (
                questionnaire_id,
                version_number,
                status
            )
            VALUES (
                target_questionnaire_id,
                next_version_number,
                'draft'
            )
            RETURNING id INTO seed_version_id;
        END IF;

        SELECT COUNT(*)
        INTO question_count
        FROM wheelset_fit_questions
        WHERE questionnaire_version_id = seed_version_id;

        IF question_count = 0 THEN
            INSERT INTO wheelset_fit_questions (
                questionnaire_version_id,
                question_key,
                answer_key,
                sort_order,
                input_mode,
                is_required,
                allow_unknown,
                is_enabled,
                source_revision
            )
            VALUES (
                seed_version_id,
                'riding_category',
                'riding_category',
                10,
                'single_choice',
                TRUE,
                TRUE,
                TRUE,
                1
            )
            RETURNING id INTO riding_question_id;

            INSERT INTO wheelset_fit_question_translations (
                question_id,
                locale,
                prompt,
                help_title,
                help_body,
                source_revision,
                translated_revision
            )
            VALUES
                (
                    riding_question_id,
                    'zh_cn',
                    '请选择您的主要骑行类型',
                    '为什么需要这个？',
                    '骑行类型会影响轮径、刹车形式和后续适配建议。',
                    1,
                    1
                ),
                (
                    riding_question_id,
                    'en',
                    'Choose your main riding category',
                    'Why we ask',
                    'The riding category affects wheel size, brake format, and fit suggestions.',
                    1,
                    1
                );

            INSERT INTO wheelset_fit_question_options (
                question_id,
                option_key,
                answer_value,
                sort_order,
                is_unknown,
                is_enabled,
                product_filter_effects,
                source_revision
            )
            VALUES
                (riding_question_id, 'mtb', 'mtb', 10, FALSE, TRUE, '{}'::jsonb, 1)
            RETURNING id INTO option_id;
            INSERT INTO wheelset_fit_question_option_translations (
                option_id,
                locale,
                label,
                description,
                source_revision,
                translated_revision
            )
            VALUES
                (option_id, 'zh_cn', '山地车', '', 1, 1),
                (option_id, 'en', 'Mountain bike', '', 1, 1);

            INSERT INTO wheelset_fit_question_options (
                question_id,
                option_key,
                answer_value,
                sort_order,
                is_unknown,
                is_enabled,
                product_filter_effects,
                source_revision
            )
            VALUES
                (riding_question_id, 'road_disc', 'road_disc', 20, FALSE, TRUE, '{}'::jsonb, 1)
            RETURNING id INTO option_id;
            INSERT INTO wheelset_fit_question_option_translations (
                option_id,
                locale,
                label,
                description,
                source_revision,
                translated_revision
            )
            VALUES
                (option_id, 'zh_cn', '公路碟刹', '', 1, 1),
                (option_id, 'en', 'Road disc brake', '', 1, 1);

            INSERT INTO wheelset_fit_question_options (
                question_id,
                option_key,
                answer_value,
                sort_order,
                is_unknown,
                is_enabled,
                product_filter_effects,
                source_revision
            )
            VALUES
                (riding_question_id, 'road_v_brake', 'road_v_brake', 30, FALSE, TRUE, '{}'::jsonb, 1)
            RETURNING id INTO option_id;
            INSERT INTO wheelset_fit_question_option_translations (
                option_id,
                locale,
                label,
                description,
                source_revision,
                translated_revision
            )
            VALUES
                (option_id, 'zh_cn', 'V 刹公路', '', 1, 1),
                (option_id, 'en', 'Road rim brake', '', 1, 1);

            INSERT INTO wheelset_fit_question_options (
                question_id,
                option_key,
                answer_value,
                sort_order,
                is_unknown,
                is_enabled,
                product_filter_effects,
                source_revision
            )
            VALUES
                (riding_question_id, 'snow', 'snow', 40, FALSE, TRUE, '{}'::jsonb, 1)
            RETURNING id INTO option_id;
            INSERT INTO wheelset_fit_question_option_translations (
                option_id,
                locale,
                label,
                description,
                source_revision,
                translated_revision
            )
            VALUES
                (option_id, 'zh_cn', '雪地车', '', 1, 1),
                (option_id, 'en', 'Snow bike', '', 1, 1);

            INSERT INTO wheelset_fit_question_options (
                question_id,
                option_key,
                answer_value,
                sort_order,
                is_unknown,
                is_enabled,
                product_filter_effects,
                source_revision
            )
            VALUES
                (riding_question_id, 'small_wheel', 'small_wheel', 50, FALSE, TRUE, '{}'::jsonb, 1)
            RETURNING id INTO option_id;
            INSERT INTO wheelset_fit_question_option_translations (
                option_id,
                locale,
                label,
                description,
                source_revision,
                translated_revision
            )
            VALUES
                (option_id, 'zh_cn', '小轮车', '', 1, 1),
                (option_id, 'en', 'Small wheel bike', '', 1, 1);

            INSERT INTO wheelset_fit_question_options (
                question_id,
                option_key,
                answer_value,
                sort_order,
                is_unknown,
                is_enabled,
                product_filter_effects,
                source_revision
            )
            VALUES
                (riding_question_id, 'unknown', 'unknown', 60, TRUE, TRUE, '{}'::jsonb, 1)
            RETURNING id INTO option_id;
            INSERT INTO wheelset_fit_question_option_translations (
                option_id,
                locale,
                label,
                description,
                source_revision,
                translated_revision
            )
            VALUES
                (option_id, 'zh_cn', '不确定', '让客服根据车型继续确认。', 1, 1),
                (option_id, 'en', 'Not sure', 'Support can confirm this from your bike model.', 1, 1);

            INSERT INTO wheelset_fit_questions (
                questionnaire_version_id,
                question_key,
                answer_key,
                sort_order,
                input_mode,
                is_required,
                allow_unknown,
                is_enabled,
                source_revision
            )
            VALUES (
                seed_version_id,
                'wheel_size',
                'wheel_size',
                20,
                'single_choice',
                TRUE,
                TRUE,
                TRUE,
                1
            )
            RETURNING id INTO wheel_size_question_id;

            INSERT INTO wheelset_fit_question_translations (
                question_id,
                locale,
                prompt,
                help_title,
                help_body,
                source_revision,
                translated_revision
            )
            VALUES
                (
                    wheel_size_question_id,
                    'zh_cn',
                    '请选择车架或前叉支持的轮径',
                    '不确定也可以继续',
                    '轮径会直接用于商品规格筛选。',
                    1,
                    1
                ),
                (
                    wheel_size_question_id,
                    'en',
                    'Choose the wheel size your frame or fork supports',
                    'You can continue if unsure',
                    'Wheel size is used directly as a product specification filter.',
                    1,
                    1
                );

            INSERT INTO wheelset_fit_question_options (
                question_id,
                option_key,
                answer_value,
                sort_order,
                is_unknown,
                is_enabled,
                product_filter_effects,
                source_revision
            )
            VALUES
                (wheel_size_question_id, 'size_20', '20', 10, FALSE, TRUE, '{"spec_filters":{"wheel_size":["20"]}}'::jsonb, 1)
            RETURNING id INTO option_id;
            INSERT INTO wheelset_fit_question_option_translations (
                option_id,
                locale,
                label,
                description,
                source_revision,
                translated_revision
            )
            VALUES
                (option_id, 'zh_cn', '20"', '', 1, 1),
                (option_id, 'en', '20"', '', 1, 1);

            INSERT INTO wheelset_fit_question_options (
                question_id,
                option_key,
                answer_value,
                sort_order,
                is_unknown,
                is_enabled,
                product_filter_effects,
                source_revision
            )
            VALUES
                (wheel_size_question_id, 'size_26', '26', 20, FALSE, TRUE, '{"spec_filters":{"wheel_size":["26"]}}'::jsonb, 1)
            RETURNING id INTO option_id;
            INSERT INTO wheelset_fit_question_option_translations (
                option_id,
                locale,
                label,
                description,
                source_revision,
                translated_revision
            )
            VALUES
                (option_id, 'zh_cn', '26"', '', 1, 1),
                (option_id, 'en', '26"', '', 1, 1);

            INSERT INTO wheelset_fit_question_options (
                question_id,
                option_key,
                answer_value,
                sort_order,
                is_unknown,
                is_enabled,
                product_filter_effects,
                source_revision
            )
            VALUES
                (wheel_size_question_id, 'size_27_5', '27.5', 30, FALSE, TRUE, '{"spec_filters":{"wheel_size":["27.5"]}}'::jsonb, 1)
            RETURNING id INTO option_id;
            INSERT INTO wheelset_fit_question_option_translations (
                option_id,
                locale,
                label,
                description,
                source_revision,
                translated_revision
            )
            VALUES
                (option_id, 'zh_cn', '27.5"', '', 1, 1),
                (option_id, 'en', '27.5"', '', 1, 1);

            INSERT INTO wheelset_fit_question_options (
                question_id,
                option_key,
                answer_value,
                sort_order,
                is_unknown,
                is_enabled,
                product_filter_effects,
                source_revision
            )
            VALUES
                (wheel_size_question_id, 'size_29', '29', 40, FALSE, TRUE, '{"spec_filters":{"wheel_size":["29"]}}'::jsonb, 1)
            RETURNING id INTO option_id;
            INSERT INTO wheelset_fit_question_option_translations (
                option_id,
                locale,
                label,
                description,
                source_revision,
                translated_revision
            )
            VALUES
                (option_id, 'zh_cn', '29"', '', 1, 1),
                (option_id, 'en', '29"', '', 1, 1);

            INSERT INTO wheelset_fit_question_options (
                question_id,
                option_key,
                answer_value,
                sort_order,
                is_unknown,
                is_enabled,
                product_filter_effects,
                source_revision
            )
            VALUES
                (wheel_size_question_id, 'size_700c', '700c', 50, FALSE, TRUE, '{"spec_filters":{"wheel_size":["700c"]}}'::jsonb, 1)
            RETURNING id INTO option_id;
            INSERT INTO wheelset_fit_question_option_translations (
                option_id,
                locale,
                label,
                description,
                source_revision,
                translated_revision
            )
            VALUES
                (option_id, 'zh_cn', '700C', '', 1, 1),
                (option_id, 'en', '700C', '', 1, 1);

            INSERT INTO wheelset_fit_question_options (
                question_id,
                option_key,
                answer_value,
                sort_order,
                is_unknown,
                is_enabled,
                product_filter_effects,
                source_revision
            )
            VALUES
                (wheel_size_question_id, 'unknown', 'unknown', 60, TRUE, TRUE, '{}'::jsonb, 1)
            RETURNING id INTO option_id;
            INSERT INTO wheelset_fit_question_option_translations (
                option_id,
                locale,
                label,
                description,
                source_revision,
                translated_revision
            )
            VALUES
                (option_id, 'zh_cn', '不确定', '让客服根据车型继续确认。', 1, 1),
                (option_id, 'en', 'Not sure', 'Support can confirm this from your bike model.', 1, 1);
        END IF;

        UPDATE wheelset_fit_questionnaire_versions
        SET
            status = 'published',
            published_at = COALESCE(published_at, NOW()),
            updated_at = NOW()
        WHERE id = seed_version_id
          AND status = 'draft';
    END IF;
END $$;

DELETE FROM selection_assistant_flows
WHERE slug = 'wheelset-fit-helper';
