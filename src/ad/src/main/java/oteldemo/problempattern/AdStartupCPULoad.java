/*
 * Copyright The OpenTelemetry Authors
 * SPDX-License-Identifier: Apache-2.0
 */
package oteldemo.problempattern;

import org.apache.logging.log4j.LogManager;
import org.apache.logging.log4j.Logger;
import java.util.Random;
import java.util.concurrent.atomic.AtomicBoolean;

/**
 * Saturates one CPU core from startup with ~300-frame deep stacks for eBPF profiler testing.
 *
 * Each ad candidate is scored through a multi-stage pipeline that recurses ~285 levels.
 * Eight distinct stage methods cycle through the stack so the profiler sees realistic,
 * named frames rather than a single repeated function. JIT inlining stops around depth 9,
 * so all deeper frames remain visible as real call stack entries.
 */
public class AdStartupCPULoad {
    private static final Logger logger = LogManager.getLogger(AdStartupCPULoad.class);
    private static final AtomicBoolean running = new AtomicBoolean(true);

    // ~285 recursion levels + ~15 JVM/thread overhead ≈ 300 total frames
    private static final int PIPELINE_DEPTH = 285;

    public static void start() {
        logger.info("Starting startup CPU load worker for profiler testing");
        Thread t = new Thread(AdStartupCPULoad::adRankingWorker, "ad-cpu-load-0");
        t.setDaemon(true);
        t.start();
    }

    // ── worker ────────────────────────────────────────────────────────────────

    private static void adRankingWorker() {
        Random rng = new Random();
        while (running.get()) {
            processAdCandidates(rng, 8);
        }
    }

    private static double processAdCandidates(Random rng, int count) {
        double total = 0;
        for (int i = 0; i < count; i++) {
            ScoringContext ctx = new ScoringContext(rng, i);
            total += runScoringPipeline(ctx, PIPELINE_DEPTH);
        }
        return total;
    }

    // ── pipeline dispatch ─────────────────────────────────────────────────────
    //
    // Eight stage methods rotate so the profiler captures distinct, named frames
    // at every level rather than a single repeated symbol.

    private static double runScoringPipeline(ScoringContext ctx, int depth) {
        if (depth <= 0) {
            return computeLeafScore(ctx);
        }
        switch (depth % 8) {
            case 0: return contextEnrichmentStage(ctx, depth);
            case 1: return featureExtractionStage(ctx, depth);
            case 2: return normalizationStage(ctx, depth);
            case 3: return modelInferenceStage(ctx, depth);
            case 4: return calibrationStage(ctx, depth);
            case 5: return filteringStage(ctx, depth);
            case 6: return aggregationStage(ctx, depth);
            default: return rankingStage(ctx, depth);
        }
    }

    // ── pipeline stages ───────────────────────────────────────────────────────

    private static double contextEnrichmentStage(ScoringContext ctx, int depth) {
        ctx.score += Math.log1p(ctx.candidateId + depth) * 0.001;
        return runScoringPipeline(ctx, depth - 1);
    }

    private static double featureExtractionStage(ScoringContext ctx, int depth) {
        for (int i = 0; i < ctx.features.length; i++) {
            ctx.features[i] = Math.sin(depth * 0.1 + i);
        }
        return runScoringPipeline(ctx, depth - 1);
    }

    private static double normalizationStage(ScoringContext ctx, int depth) {
        double norm = 0;
        for (double v : ctx.features) {
            norm += v * v;
        }
        ctx.score += Math.sqrt(norm) * 0.001;
        return runScoringPipeline(ctx, depth - 1);
    }

    private static double modelInferenceStage(ScoringContext ctx, int depth) {
        double dot = 0;
        for (int i = 0; i < ctx.features.length; i++) {
            dot += ctx.features[i] * Math.cos(i * 0.3);
        }
        ctx.score += sigmoid(dot) * 0.01;
        return runScoringPipeline(ctx, depth - 1);
    }

    private static double calibrationStage(ScoringContext ctx, int depth) {
        ctx.score = ctx.score / (1.0 + Math.abs(ctx.score));
        return runScoringPipeline(ctx, depth - 1);
    }

    private static double filteringStage(ScoringContext ctx, int depth) {
        ctx.score = Math.max(0.0, Math.min(1.0, ctx.score));
        return runScoringPipeline(ctx, depth - 1);
    }

    private static double aggregationStage(ScoringContext ctx, int depth) {
        ctx.score = (ctx.score + ctx.rng.nextDouble()) * 0.5;
        return runScoringPipeline(ctx, depth - 1);
    }

    private static double rankingStage(ScoringContext ctx, int depth) {
        ctx.score += Math.tanh(ctx.score) * 0.01;
        return runScoringPipeline(ctx, depth - 1);
    }

    // ── leaf computation ──────────────────────────────────────────────────────

    private static double computeLeafScore(ScoringContext ctx) {
        double norm = 0;
        for (double v : ctx.features) {
            norm += v * v;
        }
        return ctx.score + Math.sqrt(norm);
    }

    // ── helpers ───────────────────────────────────────────────────────────────

    private static double sigmoid(double x) {
        return 1.0 / (1.0 + Math.exp(-x));
    }

    // ── context ───────────────────────────────────────────────────────────────

    private static final class ScoringContext {
        final Random rng;
        final int candidateId;
        double score;
        final double[] features;

        ScoringContext(Random rng, int candidateId) {
            this.rng = rng;
            this.candidateId = candidateId;
            this.score = 0.0;
            this.features = new double[16];
        }
    }
}
