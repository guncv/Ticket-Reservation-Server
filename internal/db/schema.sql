\restrict 42Bgtfq38H9RVuuWPFHReEZ8W9TDeadg5tbdwUYVKK7tcwvEFfkPNJtH3fLeFoP

-- Dumped from database version 18.1
-- Dumped by pg_dump version 18.1 (Homebrew)

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET transaction_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: schema_migrations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.schema_migrations (
    version character varying NOT NULL
);


--
-- Name: schema_seeds; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.schema_seeds (
    name character varying(255) NOT NULL,
    executed_at timestamp without time zone DEFAULT now() NOT NULL
);


--
-- Name: trip_versions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.trip_versions (
    id uuid DEFAULT uuidv7() NOT NULL,
    trip_id uuid NOT NULL,
    parent_version_id uuid,
    title character varying(255) NOT NULL,
    started_date date NOT NULL,
    ended_date date NOT NULL,
    status text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    created_by_type text NOT NULL,
    CONSTRAINT trip_versions_check CHECK ((started_date <= ended_date)),
    CONSTRAINT trip_versions_created_by_type_check CHECK ((created_by_type = ANY (ARRAY['user'::text, 'ai'::text]))),
    CONSTRAINT trip_versions_status_check CHECK ((status = ANY (ARRAY['generating'::text, 'completed'::text])))
);


--
-- Name: trips; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.trips (
    id uuid DEFAULT uuidv7() NOT NULL,
    user_id uuid NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone
);


--
-- Name: user_auth_provider; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.user_auth_provider (
    id uuid DEFAULT uuidv7() NOT NULL,
    user_id uuid NOT NULL,
    provider character varying(20) NOT NULL,
    provider_id character varying(255) NOT NULL
);


--
-- Name: user_email_verify_tokens; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.user_email_verify_tokens (
    id uuid DEFAULT uuidv7() NOT NULL,
    user_id uuid NOT NULL,
    token_hash character varying(255) NOT NULL,
    expires_at timestamp with time zone NOT NULL
);


--
-- Name: user_password_reset_tokens; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.user_password_reset_tokens (
    id uuid DEFAULT uuidv7() NOT NULL,
    user_id uuid NOT NULL,
    token_hash character varying(255) NOT NULL,
    expires_at timestamp with time zone NOT NULL,
    used_at timestamp with time zone
);


--
-- Name: user_session; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.user_session (
    id uuid DEFAULT uuidv7() NOT NULL,
    user_id uuid NOT NULL,
    refresh_token_hash character varying(255) NOT NULL,
    is_revoked boolean DEFAULT false NOT NULL,
    user_agent character varying(255) NOT NULL,
    ip_address character varying(255) NOT NULL,
    revoked_at timestamp with time zone,
    expires_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: users; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.users (
    id uuid DEFAULT uuidv7() NOT NULL,
    email character varying(255) NOT NULL,
    password_hash character varying(255),
    is_email_verified boolean DEFAULT false NOT NULL,
    full_name character varying(255),
    first_name character varying(255),
    last_name character varying(255),
    profile_url character varying(255),
    locale character varying(10),
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: schema_migrations schema_migrations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.schema_migrations
    ADD CONSTRAINT schema_migrations_pkey PRIMARY KEY (version);


--
-- Name: schema_seeds schema_seeds_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.schema_seeds
    ADD CONSTRAINT schema_seeds_pkey PRIMARY KEY (name);


--
-- Name: trip_versions trip_versions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.trip_versions
    ADD CONSTRAINT trip_versions_pkey PRIMARY KEY (id);


--
-- Name: trips trips_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.trips
    ADD CONSTRAINT trips_pkey PRIMARY KEY (id);


--
-- Name: user_auth_provider user_auth_provider_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_auth_provider
    ADD CONSTRAINT user_auth_provider_pkey PRIMARY KEY (id);


--
-- Name: user_auth_provider user_auth_provider_user_id_provider_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_auth_provider
    ADD CONSTRAINT user_auth_provider_user_id_provider_key UNIQUE (user_id, provider);


--
-- Name: user_email_verify_tokens user_email_verify_tokens_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_email_verify_tokens
    ADD CONSTRAINT user_email_verify_tokens_pkey PRIMARY KEY (id);


--
-- Name: user_password_reset_tokens user_password_reset_tokens_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_password_reset_tokens
    ADD CONSTRAINT user_password_reset_tokens_pkey PRIMARY KEY (id);


--
-- Name: user_session user_session_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_session
    ADD CONSTRAINT user_session_pkey PRIMARY KEY (id);


--
-- Name: users users_email_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_email_key UNIQUE (email);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- Name: trip_versions trip_versions_parent_version_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.trip_versions
    ADD CONSTRAINT trip_versions_parent_version_id_fkey FOREIGN KEY (parent_version_id) REFERENCES public.trip_versions(id) ON DELETE CASCADE;


--
-- Name: trip_versions trip_versions_trip_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.trip_versions
    ADD CONSTRAINT trip_versions_trip_id_fkey FOREIGN KEY (trip_id) REFERENCES public.trips(id) ON DELETE CASCADE;


--
-- Name: trips trips_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.trips
    ADD CONSTRAINT trips_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: user_auth_provider user_auth_provider_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_auth_provider
    ADD CONSTRAINT user_auth_provider_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: user_email_verify_tokens user_email_verify_tokens_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_email_verify_tokens
    ADD CONSTRAINT user_email_verify_tokens_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: user_password_reset_tokens user_password_reset_tokens_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_password_reset_tokens
    ADD CONSTRAINT user_password_reset_tokens_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: user_session user_session_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_session
    ADD CONSTRAINT user_session_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- PostgreSQL database dump complete
--

\unrestrict 42Bgtfq38H9RVuuWPFHReEZ8W9TDeadg5tbdwUYVKK7tcwvEFfkPNJtH3fLeFoP


--
-- Dbmate schema migrations
--

INSERT INTO public.schema_migrations (version) VALUES
    ('20251223172120'),
    ('20251224160851'),
    ('20260102075747');
